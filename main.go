package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	_ "os"
)

func main() {
	// Настройки для Telegram-бота
	botToken := "APIKEY"
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	bot.Debug = true

	// Настройки для SSH-подключения
	sshUser := "LOGIN"
	sshHost := "HOST"
	sshPort := "22"
	privateKeyPath := "PATH"

	fmt.Print("Enter passphrase: ")
	passphrase := "java"

	// Чтение приватного ключа для SSH
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // Если есть новое сообщение
			// Подключаемся по SSH и выполняем команду
			output, err := runSSHCommand(sshConfig, sshHost+":"+sshPort, "docker ps --format '{{.Names}}'")
			if err != nil {
				log.Printf("Failed to run SSH command: %v", err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: Unable to fetch Docker containers.")
				_, err := bot.Send(msg)
				if err != nil {
					return
				}
				continue
			}

			// Отправляем результат пользователю
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Docker Containers:\n%s", output))
			_, err = bot.Send(msg)
			if err != nil {
				return
			}
		}
	}
}

func runSSHCommand(config *ssh.ClientConfig, server string, cmd string) (string, error) {
	conn, err := ssh.Dial("tcp", server, config)
	if err != nil {
		return "", fmt.Errorf("unable to connect: %v", err)
	}
	defer func(conn *ssh.Client) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("unable to create session: %v", err)
	}
	defer func(session *ssh.Session) {
		err := session.Close()
		if err != nil {

		}
	}(session)

	output, err := session.Output(cmd)
	if err != nil {
		return "", fmt.Errorf("unable to run command: %v", err)
	}

	return string(output), nil
}
