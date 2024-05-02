package files

import (
	"ReadLaterBot/storage"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Storage struct {
	basePath string
}

const (
	defaultPerm = 0774
)

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s *Storage) Save(page *storage.Page) error {
	fPath := filepath.Join(s.basePath, page.UserName)

	if err := os.MkdirAll(fPath, defaultPerm); err != nil {
		return fmt.Errorf("can't create folder: %w", err)
	}

	fName, err := fileName(page)
	if err != nil {
		return err
	}
	fPath = filepath.Join(fPath, fName)

	file, err := os.Create(fPath)
	if err != nil {
		return err
	}

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	file.Close()
	return nil
}

func (s *Storage) PickRandom(userName string) (*storage.Page, error) {
	fPath := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(fPath)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("no saved pages")
	}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(files))

	file := files[n]

	return s.decodePage(filepath.Join(fPath, file.Name()))
}

func (s Storage) Remove(p *storage.Page) error {
	fileName, err := fileName(p)
	if err != nil {
		return err
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err := os.Remove(path); err != nil {
		return errors.New(fmt.Sprintf("can't remove file %s", path))
	}

	return nil
}

func (s Storage) decodePage(filePath string) (*storage.Page, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, err
	}
	return &p, err

	f.Close()
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}
