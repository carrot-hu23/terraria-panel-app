package server

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Game struct {
	lock       sync.Mutex
	running    atomic.Bool
	send       chan []string
	binPath    string
	configPath string
	stdin      io.WriteCloser
	stdout     io.ReadCloser
}

type BackupInfo struct {
	CreateTime time.Time `json:"createTime"`
	FileName   string    `json:"fileName"`
	FileSize   int64     `json:"fileSize"`
	Time       int64     `json:"time"`
	Path       string    `json:"path"`
}

const tLogsTxt = "t_log.txt"

func NewGame(binPath, configPath string) *Game {
	running := atomic.Bool{}
	running.Store(false)
	game := &Game{
		lock:       sync.Mutex{},
		running:    running,
		send:       make(chan []string),
		binPath:    binPath,
		configPath: configPath,
	}
	return game
}

func (receiver *Game) Status() bool {
	return receiver.running.Load()
}

func (receiver *Game) Start() {
	if receiver.running.Load() == true {
		return
	}
	receiver.lock.Lock()
	// 创建输出文件
	logFile, err := os.Create(tLogsTxt)
	if err != nil {
		log.Println("Error creating log file:", err)
		return
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			receiver.lock.Unlock()
		}
	}(logFile)

	// 创建一个 cmd 对象
	cmd := exec.Command(receiver.binPath, "-config", receiver.configPath)
	receiver.running.Store(true)
	receiver.lock.Unlock()
	// 获取子进程的 stdin、stdout 和 stderr
	receiver.stdin, err = cmd.StdinPipe()
	if err != nil {
		log.Printf("Error getting stdin pipe: %v\n", err)
		return
	}
	receiver.stdout, err = cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error getting stdout pipe: %v\n", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error getting stderr pipe: %v\n", err)
		return
	}

	// 启动子进程
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		return
	}

	// 开启 goroutine 实时读取子进程的输出并写入文件和控制台
	go func() {
		io.Copy(io.MultiWriter(os.Stdout, logFile), receiver.stdout)
	}()
	go func() {
		io.Copy(io.MultiWriter(os.Stderr, logFile), stderr)
	}()

	// 等待子进程退出
	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command: %v\n", err)
		receiver.running.Store(false)
	}
	log.Println("Terraria process exit !!!")
	receiver.running.Store(false)
}

func (receiver *Game) Stop() {
	receiver.lock.Lock()
	defer receiver.lock.Unlock()

	if receiver.running.Load() == true {
		err := receiver.Send("exit")
		if err != nil {
			log.Println("stop game error", err)
		} else {
			receiver.running.Store(false)
		}
	}
}

func (receiver *Game) Send(cmd string) error {
	// 向子进程写入命令
	writer := bufio.NewWriter(receiver.stdin)
	input := cmd + "\n"
	_, err := writer.WriteString(input)
	if err != nil {
		log.Printf("Error writing to stdin: %v\n", err)
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (receiver *Game) Logs(lineNum uint) ([]string, error) {
	file, err := os.Open(tLogsTxt)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	//获取文件大小
	fs, err := file.Stat()
	fileSize := fs.Size()

	var offset int64 = -1   //偏移量，初始化为-1，若为0则会读到EOF
	char := make([]byte, 1) //用于读取单个字节
	lineStr := ""           //存放一行的数据
	buff := make([]string, 0, 100)
	for (-offset) <= fileSize {
		//通过Seek函数从末尾移动游标然后每次读取一个字节
		file.Seek(offset, io.SeekEnd)
		_, err := file.Read(char)
		if err != nil {
			return buff, err
		}
		if char[0] == '\n' {
			// offset--  //windows跳过'\r'
			lineNum-- //到此读取完一行
			buff = append(buff, lineStr)
			lineStr = ""
			if lineNum == 0 {
				return buff, nil
			}
		} else {
			lineStr = string(char) + lineStr
		}
		offset--
	}
	buff = append(buff, lineStr)
	return buff, nil
}

func (receiver *Game) GetConfig() (string, error) {
	data, err := os.ReadFile(receiver.configPath)
	if err != nil {
		fmt.Println("File reading error: ", err)
		return "", err
	}
	return string(data), err
}

func (receiver *Game) EditConfig(config string) error {
	filename := receiver.configPath
	// 判断文件是否存在
	var file *os.File
	if _, err := os.Stat(filename); os.IsNotExist(err) {

		file, err = os.Create(filename)
		if err != nil {
			fmt.Println(err)
		}

	} else {
		//O_APPEND
		file, err = os.OpenFile(filename, os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	w := bufio.NewWriter(file)
	_, err2 := w.WriteString(config)
	if err2 != nil {
		return err2
	}
	err := w.Flush()
	if err != nil {
		return err
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (receiver *Game) GetWorld() string {
	config, err := receiver.GetConfig()
	if err != nil {
		return ""
	}
	split := strings.Split(config, "\n")
	for i := range split {
		if strings.Contains(split[i], "world=") {
			lines := strings.Split(split[i], "world=")
			if len(lines) == 2 {
				return lines[1]
			}
		}
	}
	return ""
}

func (receiver *Game) GetBackupList() []BackupInfo {
	var backupList []BackupInfo
	//获取文件或目录相关信息
	dir := filepath.Dir(receiver.GetWorld())
	log.Println("backup path: ", dir)
	fileInfoList, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Panicln(err)
	}
	for _, file := range fileInfoList {
		if file.IsDir() {
			continue
		}
		suffix := filepath.Ext(file.Name())
		if strings.Contains(suffix, "bak") {
			backup := BackupInfo{
				FileName:   file.Name(),
				FileSize:   file.Size(),
				CreateTime: file.ModTime(),
				Time:       file.ModTime().Unix(),
				Path:       filepath.Join(dir, file.Name()),
			}
			backupList = append(backupList, backup)
		}
	}
	return backupList
}

func (receiver *Game) Restore(backupFilePath string) error {
	worldPath := receiver.GetWorld()
	// 1. 删除文件A
	err := os.Remove(worldPath)
	if err != nil {
		return err
	}
	// 2. 将文件B重命名为文件A
	err = os.Rename(backupFilePath, worldPath)
	return err
}

func (receiver *Game) DeleteBackup(backupFilePath string) error {

	err := os.Remove(backupFilePath)
	return err

}
