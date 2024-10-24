package server

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/mnako/letters"
)

type StoredFile struct {
	UserId   int64
	FileName string
}

func (file *StoredFile) Filename() string {
	return fmt.Sprintf("mail/%d/%s", file.UserId, file.FileName)
}

func (file *StoredFile) Exists() bool {
	_, err := os.Stat(file.Filename())
	return !os.IsNotExist(err)
}

func (file *StoredFile) WebUrl(baseUrl string, hashGenerator HashGenerator) string {
	hash := "no-hash-generator"
	if hashGenerator != nil {
		hash = hashGenerator.CreateHash(fmt.Sprintf("%d%s", file.UserId, file.FileName))
	}
	return fmt.Sprintf("%s/mail/%d/%s?hash=%s", baseUrl, file.UserId, file.FileName, hash)
}

func (file *StoredFile) createFile() (*os.File, error) {
	err := os.MkdirAll(fmt.Sprintf("mail/%d", file.UserId), 0755)
	if err != nil {
		return nil, err
	}
	fileName := file.Filename()
	return os.Create(fileName)
}

func (file *StoredFile) SaveString(data string) error {

	f, err := file.createFile()
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(data)
	return err
}

func (file *StoredFile) SaveRaw(email letters.Email) error {
	f, err := file.createFile()
	if err != nil {
		return err
	}
	defer f.Close()

	fz := gzip.NewWriter(f)
	defer fz.Close()

	enc := gob.NewEncoder(fz)
	err = enc.Encode(email)
	return err
}

func (file *StoredFile) SaveData(data []byte) error {
	f, err := file.createFile()
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

type StorageResult struct {
	Html        StoredFile
	Raw         StoredFile
	Attachments []StoredFile
}

func saveMail(emailId string, userId int64, email letters.Email) (StorageResult, error) {
	var err error
	ret := StorageResult{
		Attachments: []StoredFile{},
	}

	saveRawEmail := StoredFile{
		UserId:   userId,
		FileName: fmt.Sprintf("%s.eml", emailId),
	}
	err = saveRawEmail.SaveRaw(email)
	if err != nil {
		log.Printf("Error saving raw email: %v", err)
		ret.Raw = saveRawEmail
	}

	ret.Html = StoredFile{
		UserId:   userId,
		FileName: fmt.Sprintf("%s.html", emailId),
	}
	if email.HTML == "" {
		err = ret.Html.SaveString(fmt.Sprintf("<html><body><code><pre>%s</pre></code></body></html>", email.Text))
	} else {
		err = ret.Html.SaveString(email.HTML)
	}
	if err != nil {
		return ret, err
	}

	for i, attachment := range email.AttachedFiles {
		att := StoredFile{
			UserId:   userId,
			FileName: fmt.Sprintf("%s-%d", emailId, i),
		}
		file, ok := attachment.ContentDisposition.Params["filename"]
		if ok {
			att.FileName = file
		}
		err = att.SaveData(attachment.Data)
		if err == nil {
			ret.Attachments = append(ret.Attachments, att)
		}

	}
	return ret, err
}
