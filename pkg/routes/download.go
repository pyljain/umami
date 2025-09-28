package routes

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"umami/pkg/db"
)

func Download(database db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appId := r.PathValue("id")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		rootPath := path.Join("./repository", appId)
		buf := new(bytes.Buffer)
		archive := zip.NewWriter(buf)
		directoriesToIgnore := map[string]struct{}{
			".git":         {},
			"venv":         {},
			".venv":        {},
			"node_modules": {},
		}

		err := filepath.WalkDir(rootPath, func(filePath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			relativePath, err := filepath.Rel(rootPath, filePath)
			if err != nil {
				return err
			}

			if _, exists := directoriesToIgnore[d.Name()]; exists && d.IsDir() {
				return fs.SkipDir
			}

			if d.IsDir() {
				return nil
			}

			f, err := archive.Create(relativePath)
			if err != nil {
				log.Printf("Unable to add file %s to archive: %s", relativePath, err)
				return err
			}

			fileBytes, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Unable to read file %s to archive: %s", relativePath, err)
				return err
			}

			_, err = f.Write(fileBytes)
			if err != nil {
				log.Printf("Unable to write file %s to archive: %s", relativePath, err)
				return err
			}

			return nil

		})

		if err != nil {
			log.Printf("Unable to generate archive %s", err)
			http.Error(w, "Unable to generate archive", http.StatusInternalServerError)
			return
		}

		err = archive.Close()
		if err != nil {
			log.Printf("Unable to generate archive %s", err)
			http.Error(w, "Unable to generate archive", http.StatusInternalServerError)
			return
		}

		w.Write(buf.Bytes())
	}
}
