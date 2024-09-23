package main

import (
	"fmt"
	"log"
	"runtime"
	"terraria-panel/api"
	"terraria-panel/internal/config"
	"terraria-panel/internal/global"
	"terraria-panel/server"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	version = "1.1.0"
	cfgFile string
	conf    config.Config
)

func main() {
	config.Init(cfgFile, &conf)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("version", "1.1.0")
		c.Next()
	})
	api.RegisterRouter(router)

	go func() {
		fmt.Printf("Current OS: %s\n", runtime.GOOS)
		var binPath string
		var configPath string
		if runtime.GOOS == "windows" {
			binPath = ".\\Terraria-1449\\Linux\\\\TerrariaServer.exe"
			configPath = ".\\Terraria-1449\\Linux\\config.txt"
		} else {
			binPath = "./Terraria-1449/Linux/TerrariaServer.bin.x86_64"
			configPath = "./config.txt"
		}
		// global.TerrariaGame = server.NewGame(`C:\\Users\\paratera\\Desktop\\我的\\泰拉瑞亚\\1449\\Linux\\TerrariaServer.exe`, `C:\Users\paratera\Desktop\我的\泰拉瑞亚\1449\Linux\config.txt`)
		global.TerrariaGame = server.NewGame(binPath, configPath)
	}()

	log.Println("Starting terraria-panel...")
	log.Println("Version: ", version)
	log.Println("Port:", viper.GetInt("web.port"))

	if err := router.Run(fmt.Sprintf(":%d", viper.GetInt("web.port"))); err != nil {
		log.Panicln("Server exited with error: ", err)
	}

	//// 创建输出文件
	//logFile, err := os.Create("t_log.txt")
	//if err != nil {
	//	fmt.Println("Error creating log file:", err)
	//	return
	//}
	//defer logFile.Close()
	//
	//// 创建一个 cmd 对象
	//cmd := exec.Command(`C:\Users\paratera\Desktop\我的\泰拉瑞亚\1449\Linux\TerrariaServer.exe`, `-config`, `C:\Users\paratera\Desktop\我的\泰拉瑞亚\1449\Linux\config.txt`)
	//
	//// 获取子进程的 stdin、stdout 和 stderr
	//stdin, err := cmd.StdinPipe()
	//if err != nil {
	//	fmt.Printf("Error getting stdin pipe: %v\n", err)
	//	return
	//}
	//stdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	fmt.Printf("Error getting stdout pipe: %v\n", err)
	//	return
	//}
	//stderr, err := cmd.StderrPipe()
	//if err != nil {
	//	fmt.Printf("Error getting stderr pipe: %v\n", err)
	//	return
	//}
	//
	//// 启动子进程
	//if err := cmd.Start(); err != nil {
	//	fmt.Printf("Error starting command: %v\n", err)
	//	return
	//}
	//
	//// // 开启 goroutine 实时读取子进程的输出
	//// go func() {
	//// 	io.Copy(os.Stdout, stdout)
	//// }()
	//// go func() {
	//// 	io.Copy(os.Stderr, stderr)
	//// }()
	//
	//// 开启 goroutine 实时读取子进程的输出并写入文件和控制台
	//go func() {
	//	io.Copy(io.MultiWriter(os.Stdout, logFile), stdout)
	//}()
	//go func() {
	//	io.Copy(io.MultiWriter(os.Stderr, logFile), stderr)
	//}()
	//
	//// 向子进程写入命令
	//writer := bufio.NewWriter(stdin)
	//for {
	//	// 读取用户输入
	//	reader := bufio.NewReader(os.Stdin)
	//	input, _ := reader.ReadString('\n')
	//	// input = input[:len(input)-1] // 去掉换行符
	//
	//	// 如果输入 'q'，则退出循环
	//	if input == "q" {
	//		break
	//	}
	//	_, err := writer.WriteString(input)
	//	if err != nil {
	//		fmt.Printf("Error writing to stdin: %v\n", err)
	//		return
	//	}
	//	writer.Flush()
	//}
	//
	//// 等待子进程退出
	//if err := cmd.Wait(); err != nil {
	//	fmt.Printf("Error waiting for command: %v\n", err)
	//}
	//
	//fmt.Println("子进程已退出")
}
