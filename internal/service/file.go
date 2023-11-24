package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var headers = [...]string{"n", "mqqt", "invid", "unit_guid", "msg_id", "text", "context", "class", "level", "area", "addr", "block", "type", "bit", "invert_bit"}

type TargetFile struct {
	Name    string   `bson:"_id"`
	Records []Record `bson:"records"`
	Errors  []string `bson:"errors"`
}

type Record struct {
	N          string `bson:"n"`
	Mqqt       string `bson:"mqqt"`
	Invid      string `bson:"invid"`
	Unit_guid  string `bson:"guid"`
	Msg_id     string `bson:"msg_id"`
	Text       string `bson:"text"`
	Context    string `bson:"context"`
	Class      string `bson:"class"`
	Level      string `bson:"level"`
	Area       string `bson:"area"`
	Addr       string `bson:"addr"`
	Block      string `bson:"block"`
	Type       string `bson:"type"`
	Bit        string `bson:"bit"`
	Invert_bit string `bson:"invert_bit"`
}

func readTsv(file *os.File) TargetFile {
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	tf := &TargetFile{
		Records: make([]Record, 0),
		Errors:  make([]string, 0),
	}

	var index int
	var line []string
	var err error
	for {
		line, err = reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			tf.Errors = append(tf.Errors, err.Error())
			index++
			continue
		}
		if index > 1 {
			rec, errs := parseTsv(line, index)
			tf.Records = append(tf.Records, *rec)
			tf.Errors = append(tf.Errors, errs...)
		}
		index++
	}
	if index == 0 {
		tf.Errors = append(tf.Errors, "Not enough header lines in file (2 required)")
	}
	return *tf
}

func parseTsv(line []string, index int) (*Record, []string) {
	rec := &Record{}
	errors := make([]string, 0, len(line))
	if len(line) != len(headers) {
		errors = append(errors, fmt.Sprintf("line %d, nvalid number of arguments: %d (must be %d)", index, len(line), len(headers)))
	}
	for i, v := range line {
		v = strings.TrimSpace(v)
		switch i {
		case 0:
			if v != "" {
				_, err := strconv.Atoi(v)
				if err != nil {
					errors = append(errors, fmt.Sprintf("line %d, field \"%s\" must be a number", index, headers[i]))
					v = ""
				}
			}
			rec.N = v
		case 1:
			rec.Mqqt = v
		case 2:
			rec.Invid = v
		case 3:
			rec.Unit_guid = v
		case 4:
			rec.Msg_id = v
		case 5:
			rec.Text = v
		case 6:
			rec.Context = v
		case 7:
			rec.Class = v
		case 8:
			if v != "" {
				_, err := strconv.Atoi(v)
				if err != nil {
					errors = append(errors, fmt.Sprintf("line %d, field \"%s\" must be a number", index, headers[i]))
					v = ""
				}
			}
			rec.Level = v
		case 9:
			rec.Area = v
		case 10:
			rec.Addr = v
		case 11:
			rec.Block = v
		case 12:
			rec.Type = v
		case 13:
			if v != "" {
				_, err := strconv.Atoi(v)
				if err != nil {
					errors = append(errors, fmt.Sprintf("line %d, field \"%s\" must be a number", index, headers[i]))
					v = ""
				}
			}
			rec.Bit = v
		case 14:
			if v != "" {
				_, err := strconv.Atoi(v)
				if err != nil {
					errors = append(errors, fmt.Sprintf("line %d, field \"%s\" must be a number", index, headers[i]))
					v = ""
				}
			}
			rec.Invert_bit = v
		}
	}
	return rec, errors
}
