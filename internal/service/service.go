package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"slices"
	"text/tabwriter"
	"text/template"
	"time"

	"log/slog"

	"github.com/signintech/gopdf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	Db             *mongo.Database
	Logger         *slog.Logger
	CollectionName string
}

func (s *Service) WriteFilesToDB(files []*TargetFile) error {
	if len(files) == 0 {
		s.Logger.InfoContext(context.Background(), "no new files")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var mrsh []interface{}
	for _, v := range files {
		bytes, err := bson.Marshal(v)
		if err != nil {
			return err
		}
		mrsh = append(mrsh, bytes)
	}
	_, err := s.Db.Collection(s.CollectionName).InsertMany(ctx, mrsh)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetProcessedFileNames() ([]string, error) {
	var res []string
	var objs []bson.D
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cur, err := s.Db.Collection(s.CollectionName).Find(
		ctx,
		bson.D{},
		options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	if err = cur.All(ctx, &objs); err != nil {
		return nil, err
	}
	for _, v := range objs {
		id, ok := v[0].Value.(string)
		if ok {
			res = append(res, id)
		}
	}
	s.Logger.InfoContext(ctx, "filenames retreived", slog.Int("already processed", len(objs)))
	return res, nil
}

func (s *Service) GetDocs(page, limit int) ([]*TargetFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	skip := int64(limit * (page - 1))
	lim := int64(limit)

	res := make([]*TargetFile, 0, limit)
	cur, err := s.Db.Collection(s.CollectionName).Find(
		ctx,
		bson.D{},
		&options.FindOptions{
			Limit: &lim,
			Skip:  &skip,
		},
	)
	if err != nil {
		return nil, err
	}

	for cur.Next(ctx) {
		var tf *TargetFile
		if err := cur.Decode(&tf); err != nil {
			fmt.Println(err)
			return nil, err
		}
		res = append(res, tf)
	}
	return res, nil
}

func (s *Service) RunDirectiryScanner(fileDir, outputDir string) error {
	ctx := context.Background()
	if err := os.MkdirAll(fileDir, 0744); err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0744); err != nil {
		return err
	}
	ticker := time.Tick(20 * time.Second)
	for {
		select {
		case <-ticker:
			alreadyProcessed, err := s.GetProcessedFileNames()
			if err != nil {
				s.Logger.ErrorContext(ctx, err.Error())
				break
			}
			processedFiles, err := s.processDir(fileDir, outputDir, alreadyProcessed)
			if err != nil {
				s.Logger.ErrorContext(ctx, err.Error())
				break
			}
			if err = s.WriteFilesToDB(processedFiles); err != nil {
				s.Logger.ErrorContext(ctx, err.Error())
			}
		}
	}
}

func (s *Service) processDir(fileDir, outputDir string, alreadyProcessed []string) ([]*TargetFile, error) {
	ctx := context.Background()
	var tsvs []*TargetFile
	entries, err := os.ReadDir(fileDir)
	if err != nil {
		return nil, err
	}
	for _, v := range entries {
		if v.IsDir() {
			continue
		}
		filePath := path.Join(fileDir, v.Name())
		if path.Ext(filePath) != ".tsv" {
			continue
		}
		if slices.Contains(alreadyProcessed, v.Name()) {
			continue
		}
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0766)
		if err != nil {
			s.Logger.ErrorContext(ctx, err.Error())
			continue
		}
		tf, toGenerate := readTsv(file)

		if err := s.createPDFs(toGenerate, outputDir); err != nil {
			s.Logger.ErrorContext(ctx, "failed to create pdf", slog.Any("err", err.Error()))
		}
		if err := s.createTxts(toGenerate, outputDir); err != nil {
			s.Logger.ErrorContext(ctx, "failed to create txt", slog.Any("err", err.Error()))
		}

		tf.Name = v.Name()
		tsvs = append(tsvs, tf)
		s.Logger.InfoContext(ctx, "file processed", slog.Any("name", tf.Name))
	}
	return tsvs, nil
}

func (s *Service) createPDFs(files map[string]*TargetFile, outDir string) error {
	if len(files) == 0 {
		s.Logger.InfoContext(context.Background(), "No new pdf files created")
		return nil
	}
	templ := "{{range .Records}}{{.N}}\t{{.Mqqt}}\t{{.Invid}}\t{{.Msg_id}}\t{{.Text}}\t{{.Context}}\t{{.Class}}\t{{.Level}}\t{{.Area}}\t{{.Addr}}\t{{.Block}}\t{{.Type}}\t{{.Bit}}\t{{.Invert_bit}}\n{{end}}"
	for _, v := range files {
		var bb bytes.Buffer
		t := template.Must(template.New("").Parse(templ))
		w := tabwriter.NewWriter(&bb, 12, 4, 2, ' ', 0)
		err := t.Execute(w, v)
		if err != nil {
			return err
		}
		w.Flush()

		pdf := gopdf.GoPdf{}
		pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA2})
		pdf.AddPage()
		pdf.AddTTFFont("Arial Cyr", "fonts/Arial Cyr.ttf")
		pdf.SetFont("Arial Cyr", "", 12)
		pdf.SetXY(10, 10)
		pdf.Text("Records")
		for ind := range v.Records {
			pdf.SetXY(10, float64(10+10*(ind+1)))
			str, _ := bb.ReadString('\n')
			pdf.Text(str)
		}

		pdf.AddPage()
		pdf.SetXY(10, 10)
		pdf.Text("Errors")
		for ind, v := range v.Errors {
			pdf.SetXY(10, float64(10+10*(ind+1)))
			pdf.Text(v)
		}
		if err = pdf.WritePdf(fmt.Sprintf("%s/%s.pdf", outDir, v.Name)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) createTxts(files map[string]*TargetFile, outDir string) error {
	if len(files) == 0 {
		s.Logger.InfoContext(context.Background(), "No new pdf files created")
		return nil
	}
	templ := "N\tMqqt\tInvid\tMsg_id\tText\tContext\tClass\tLevel\tArea\tAddr\tBlock\tType\tBit\tInvert_bit\n" +
		"{{range .Records}}{{.N}}\t{{.Mqqt}}\t{{.Invid}}\t{{.Msg_id}}\t{{.Text}}\t{{.Context}}\t{{.Class}}\t{{.Level}}\t{{.Area}}\t{{.Addr}}\t{{.Block}}\t{{.Type}}\t{{.Bit}}\t{{.Invert_bit}}\n{{end}}\n\n" +
		"Errors" +
		"{{range .Errors}}{{.}}\n{{end}}"
	for _, v := range files {
		file, err := os.Create(fmt.Sprintf("%s/%s.txt", outDir, v.Name))
		if err != nil {
			return err
		}
		t := template.Must(template.New("").Parse(templ))
		w := tabwriter.NewWriter(file, 12, 4, 2, ' ', 0)
		err = t.Execute(w, v)
		if err != nil {
			return err
		}
		w.Flush()
	}
	return nil
}
