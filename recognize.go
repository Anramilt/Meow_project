package main

import (
	"encoding/json"
	"log"
	"math"
	"sort"
	"strings"

	vosk "github.com/alphacep/vosk-api/go"
)

const frameLength = 256

var model *vosk.VoskModel

func initVosk() error {
	var err error
	model, err = vosk.NewModel("model_ru")
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func recWithGrammar(grammar []string, sampleRate int) (*vosk.VoskRecognizer, error) {
	var fixedGrammar []string
	for i := range grammar {
		fixedGrammar = append(fixedGrammar, strings.Trim(strings.ToLower(grammar[i]), " "))
	}
	marshalledGrammar, err := json.Marshal(fixedGrammar)
	if err != nil {
		log.Println(err)
	}
	rec, err := vosk.NewRecognizerGrm(model, float64(sampleRate), string(marshalledGrammar))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rec.SetWords(0)
	return rec, nil
}

type ResultCategory int

const (
	ResultBad = iota
	ResultOk
	ResultGood
)

type RecognitionResult struct {
	Value    float64
	Category ResultCategory
	Text     string
}

type GopResult struct {
	Gop  []float64 `json:"gop"`
	Text string    `json:"text"`
}

const sampleRate = 16000

func Recognize(grammar []string, data []byte) []RecognitionResult {
	var res []RecognitionResult
	rec, err := recWithGrammar(grammar, sampleRate)
	if err != nil {
		return nil
	}
	for i := 0; i < len(data)-1025; i += 1024 {
		if rec.AcceptWaveform(data[i:i+1024]) != 0 {
			result := string(rec.Result())
			if strings.Contains(string(result), "\"text\" : \"\"") {
				continue
			}
			if !strings.Contains(string(result), "gop") {
				log.Println("Vosk не той версии")
				continue
			}

			var resStruct GopResult
			json.Unmarshal([]byte(result), &resStruct)
			if len(resStruct.Gop) < 1 {
				continue
			}
			sort.Slice(resStruct.Gop, func(i, j int) bool { return resStruct.Gop[i] < resStruct.Gop[j] })
			var average float64
			for _, g := range resStruct.Gop {
				average += g
			}
			average /= float64(len(resStruct.Gop))

			value := math.Exp((resStruct.Gop[len(resStruct.Gop)/2] + resStruct.Gop[0] + average) / 3)
			var category ResultCategory
			if value < 0.6 {
				category = ResultBad
			} else if value > 0.8 {
				category = ResultGood
			} else {
				category = ResultOk
			}
			res = append(res, RecognitionResult{
				Value:    value,
				Category: category,
				Text:     resStruct.Text,
			})

		}
	}

	return res
}
