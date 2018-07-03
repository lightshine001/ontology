package utils

import (
	"bufio"
	"github.com/ontio/ontology/common/log"
	"io"
	"os"
)

// read white list or black list
func LoadList(fileName string) []string {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("load dht list file %s failed", fileName)
		return []string{}
	}

	list := make([]string, 0)
	bufferReader := bufio.NewReader(file)
	for {
		line, _, err := bufferReader.ReadLine()
		if err == io.EOF {
			break
		}
		list = append(list, string(line))
	}
	return list
}

func IsContained(list []string, destString string) bool {
	for _, element := range list {
		if element == destString {
			return true
		}
	}
	return false
}

func SaveListToFile(list []string, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Errorf("create list file %s failed!", fileName)
		return
	}
	saveContent := ""
	for _, addr := range list {
		saveContent += addr + "\n"
	}
	saveContent = saveContent[:len(saveContent)-1] // remove last \n
	n, err := file.WriteString(saveContent)
	if n != len(saveContent) || err != nil {
		log.Errorf("save list to file %s failed!", fileName)
	}
}
