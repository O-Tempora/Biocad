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
	Records []Record `bson:"records,omitempty"`
	Errors  []string `bson:"errors,omitempty"`
}

type Record struct {
	N          string   `bson:"n,omitempty"`
	Mqqt       string   `bson:"mqqt,omitempty"`
	Invid      string   `bson:"invid,omitempty"`
	Unit_guid  string   `bson:"guid,omitempty"`
	Msg_id     string   `bson:"msg_id,omitempty"`
	Text       string   `bson:"text,omitempty"`
	Context    string   `bson:"context,omitempty"`
	Class      string   `bson:"class,omitempty"`
	Level      string   `bson:"level,omitempty"`
	Area       string   `bson:"area,omitempty"`
	Addr       string   `bson:"addr,omitempty"`
	Block      string   `bson:"block,omitempty"`
	Type       string   `bson:"type,omitempty"`
	Bit        string   `bson:"bit,omitempty"`
	Invert_bit string   `bson:"invert_bit,omitempty"`
	Errors     []string `bson:"errors,omitempty"`
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
			rec := parseTsv(line)
			tf.Records = append(tf.Records, *rec)
		}
		index++
	}
	if index == 0 {
		tf.Errors = append(tf.Errors, "Not enough header lines in file (2 required)")
	}
	return *tf
}

func parseTsv(line []string) *Record {
	rec := &Record{}
	if len(line) != len(headers) {
		rec.Errors = append(rec.Errors, fmt.Sprintf("Invalid number of arguments: %d (must be %d)", len(line), len(headers)))
	}
	for i, v := range line {
		v = strings.TrimSpace(v)
		// if v == "" {
		// 	rec.Errors = append(rec.Errors, fmt.Errorf("field \"%s\" is empty", headers[i]))
		// }
		switch i {
		case 0:
			if v != "" {
				_, err := strconv.Atoi(v)
				if err != nil {
					rec.Errors = append(rec.Errors, fmt.Sprintf("field \"%s\" must be a number", headers[i]))
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
					rec.Errors = append(rec.Errors, fmt.Sprintf("field \"%s\" must be a number", headers[i]))
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
					rec.Errors = append(rec.Errors, fmt.Sprintf("field \"%s\" must be a number", headers[i]))
					v = ""
				}
			}
			rec.Bit = v
		case 14:
			if v != "" {
				_, err := strconv.Atoi(v)
				if err != nil {
					rec.Errors = append(rec.Errors, fmt.Sprintf("field \"%s\" must be a number", headers[i]))
					v = ""
				}
			}
			rec.Invert_bit = v
		}
	}
	return rec
}
