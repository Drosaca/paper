package qr

import (
	"fmt"
	rqrc "github.com/tuotoo/qrcode"
	wwqrc "github.com/yeqown/go-qrcode"
	"io/ioutil"
	"os"
	"paperify/fn"
	"paperify/stat"
	"path"
	"path/filepath"
	"strconv"
	"sync"
)

type Qr struct {
	ErrorCorrection uint // value from 0 to 3 (7%, 15%, 25%, 30%)
	InputPath       string
	OutputPath      string
	recipient       *os.File
	decodeStep      int
	chunks          map[int]string
}

func NewQr(inputPath string, outputPath string) Qr {
	return Qr{ErrorCorrection: 1, InputPath: inputPath, OutputPath: outputPath, recipient: nil, decodeStep: -1, chunks: map[int]string{}}
}

func (qr *Qr) generateQrCode(str string, filename string) error {
	fmt.Println("writing.. ", filename)
	config := wwqrc.Config{
		EcLevel: wwqrc.ErrorCorrectionLow,
		EncMode: wwqrc.EncModeByte,
	}
	qrc, err := wwqrc.NewWithConfig(str, &config)
	if err != nil {
		return err
	}
	err = qrc.Save(filename)
	return err
}

func (qr *Qr) CreateQr() error {
	f, err := os.Open(qr.InputPath)
	stats, err := f.Stat()
	if err != nil {
		return err
	}
	if stats.Size()/2951 > 255 {
		panic("file too big to be paperified Qr code files > 255")
	}
	buff := make([]byte, 2951) //b64 2213 default 2952
	part := 0
	for bytes, err := f.Read(buff); bytes != 0; {
		str := string(part) + string(buff[:bytes])
		fn.Log("size",len(str))
		_, inputFileName := path.Split(qr.InputPath)
		err = qr.generateQrCode(str, path.Join(qr.OutputPath, inputFileName+fmt.Sprintf("_%d.png", part)))
		if err != nil {
			return err
		}
		for i := range buff {
			buff[i] = 0
		}
		bytes, err = f.Read(buff)
		if err != nil {
			return err
		}
		part += 1
	}
	return err
}

func (qr *Qr) ReadQr() error {
	tmpDir, err := os.MkdirTemp("", "paper")
	if err != nil {
		return err
	}
	if stat.IsDirectory(qr.OutputPath) {
		qr.recipient, err = os.Create(filepath.Join(qr.OutputPath, "output.raw"))
	} else {
		qr.recipient, err = os.Create(qr.OutputPath)
	}
	if stat.IsDirectory(qr.InputPath) {
		err = qr.readInDir(tmpDir)
	} else {
		err := qr.decodeQr(qr.InputPath, tmpDir)
		if err != nil {
			return err
		}
	}
	err = qr.findStoredChunk()
	if err != nil {
		return err
	}
	defer qr.recipient.Close()
	defer os.RemoveAll(tmpDir)
	return err
}

func (qr *Qr) readRoutine(wg *sync.WaitGroup, path string, tmpDir string) {
	err := qr.decodeQr(path, tmpDir)
	if err != nil {
		fmt.Println(err)
	}
	defer wg.Done()
}

func (qr *Qr) readInDir(tmpDir string) error {
	var wg sync.WaitGroup

	workers := 0
	err := filepath.Walk(qr.InputPath, func(path string, info os.FileInfo, err error) error {
		if workers == 10 {
			workers = 0
			wg.Wait()
		}
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		wg.Add(1)
		go qr.readRoutine(&wg, path, tmpDir)
		workers += 1
		return err
	})
	wg.Wait()
	return err
}

func (qr *Qr) decodeQr(path string, tmpDir string) error {
	fmt.Println("reading..", path)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	qrCode, err := rqrc.Decode(file)
	if err != nil {
		return fmt.Errorf("Recognize failed: %v\n", err)
	}
	bytes := []byte(qrCode.Content)
	err = qr.writeTmpChunk(bytes, tmpDir)
	if err != nil {
		return err
	}
	return err
}

func (qr *Qr) writeTmpChunk(bytes []byte, tmpDir string) error {
	//chunkPath := path.Join(os.TempDir(), strconv.Itoa(int(bytes[0]))+".chunk")
	f, err := os.CreateTemp(tmpDir, strconv.Itoa(int(bytes[0]))+"_*.chunk" )
	chunkPath := f.Name()
	fn.Log("creating chunk", chunkPath)
	//f, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	_, err = f.Write(bytes[1:])
	if err != nil {
		return err
	}
	qr.chunks[int(bytes[0])] = chunkPath
	return err
}

func (qr *Qr) findStoredChunk() error {
	for partNumber, fPath := range qr.chunks {
		if partNumber == qr.decodeStep+1 {
			bytes, err := ioutil.ReadFile(fPath)
			if err != nil {
				return err
			}
			fn.Log("chunk found : ", fPath, " merging...")
			_, err = qr.recipient.Write(bytes[:])
			qr.decodeStep += 1
			err = qr.findStoredChunk()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
