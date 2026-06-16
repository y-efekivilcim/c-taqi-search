package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"taqi-search/engine"
)


func LoadDictionaryAsCorpus(eng *engine.Engine) error {
	file, err := os.Open("words_alpha.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	emptyArr := []string{}

	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if len(word) == 0 {
			continue
		}
		
		eng.AddDocument(word, word, emptyArr, emptyArr)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading words_alpha.txt: %v", err)
		return err
	}

	return nil
}
