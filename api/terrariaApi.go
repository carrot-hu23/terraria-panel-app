package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"terraria-panel/internal/global"
	"terraria-panel/utils/fileUtils"
	"time"
)

func getServerConfig(c *gin.Context) {
	config, err := global.TerrariaGame.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	} else {
		c.JSON(http.StatusOK, config)
	}
}

func editServerConfig(c *gin.Context) {
	var payload struct {
		Config string `json:"config"`
	}
	err := c.ShouldBind(&payload)
	if err != nil {
		log.Panicln("参数解析失败")
	}

	err = global.TerrariaGame.EditConfig(payload.Config)
	c.JSON(http.StatusOK, gin.H{})
}

func startGame(c *gin.Context) {

	go func() {
		global.TerrariaGame.Start()
	}()
	time.Sleep(5 * time.Second)
	c.JSON(http.StatusOK, gin.H{})
}

func stopGame(c *gin.Context) {
	global.TerrariaGame.Stop()
	c.JSON(http.StatusOK, gin.H{})
}

func sendCmd(c *gin.Context) {
	var payload struct {
		Cmd string `json:"cmd"`
	}
	err := c.ShouldBind(&payload)
	if err != nil {
		log.Panicln("参数解析失败")
	}
	err = global.TerrariaGame.Send(payload.Cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func gameStatus(c *gin.Context) {
	status := global.TerrariaGame.Status()
	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

func gameLogs(c *gin.Context) {
	lineNum := c.DefaultQuery("lineNum", strconv.Itoa(100))
	value, _ := strconv.ParseUint(lineNum, 10, 32)
	logs, err := global.TerrariaGame.Logs(uint(value))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	c.JSON(http.StatusOK, logs)
}

func gameBackup(c *gin.Context) {
	c.JSON(http.StatusOK, global.TerrariaGame.GetBackupList())
}

func restoreBackup(c *gin.Context) {
	backupFilePath := c.Query("backupFilePath")
	if backupFilePath == "" || !fileUtils.Exists(backupFilePath) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file path not exist"})
	}
	err := global.TerrariaGame.Restore(backupFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("restore backup error: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func deleteBackup(c *gin.Context) {
	backupFilePath := c.Query("backupFilePath")
	if backupFilePath == "" || !fileUtils.Exists(backupFilePath) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file path not exist"})
	}
	err := global.TerrariaGame.DeleteBackup(backupFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("delete backup error: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
