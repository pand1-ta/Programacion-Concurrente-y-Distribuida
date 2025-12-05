package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"sdr/api/internal/models"
	"strconv"
	"strings"
)

type MatrixData struct {
	Matrix              [][]float64
	MovieIndexToMovieID map[int]int
	MovieIDToMovieIndex map[int]int
}

func LoadMovies(path string) (map[int]models.Movie, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)

	// Leer header: movieId,title,genres
	_, err = r.Read()
	if err != nil {
		return nil, fmt.Errorf("movies csv header read: %w", err)
	}

	movies := map[int]models.Movie{}

	for {
		row, err := r.Read()
		if err != nil {
			break
		}

		id, _ := strconv.Atoi(row[0])
		title := row[1]
		genresRaw := row[2] // Adventure|Animation|Children...

		// Normalizar: minúsculas
		genres := strings.ToLower(genresRaw)

		movies[id] = models.Movie{
			MovieID: row[0],
			Title:   title,
			Genre:   genres,
		}
	}

	return movies, nil
}

func LoadMapping(path string) (map[string]int, map[int]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	if _, err := r.Read(); err != nil {
		return nil, nil, fmt.Errorf("mapping header read: %w", err)
	}
	origToIdx := map[string]int{}
	idxToOrig := map[int]string{}
	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		if row[0] == "userIndex" || row[0] == "movieIndex" {
			// row[0] is header; already skipped
		}
		// detect which order: if first is index numeric then second is id; otherwise swap
		if _, errA := strconv.Atoi(row[0]); errA == nil {
			idx, _ := strconv.Atoi(row[0])
			origToIdx[row[1]] = idx
			idxToOrig[idx] = row[1]
		} else {
			// fallback (if mapping stored as userId,userIndex)
			idx, _ := strconv.Atoi(row[1])
			origToIdx[row[0]] = idx
			idxToOrig[idx] = row[0]
		}
	}
	return origToIdx, idxToOrig, nil
}

func LoadUserMovieMatrix(path string) (*MatrixData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("matrix file empty")
	}
	header := rows[0] // header[0] == "userIndex", header[1..] are movieIndex numbers (strings)
	numMovies := len(header) - 1
	numUsers := len(rows) - 1

	movieIndexToMovieID := map[int]int{}
	movieIDToMovieIndex := map[int]int{}
	for i := 0; i < numMovies; i++ {
		id, err := strconv.Atoi(header[i+1])
		if err != nil {
			return nil, fmt.Errorf("invalid movie index in header: %v", err)
		}
		movieIndexToMovieID[i] = id
		movieIDToMovieIndex[id] = i
	}

	matrix := make([][]float64, numUsers)
	for u := 0; u < numUsers; u++ {
		row := rows[u+1]
		if len(row) != numMovies+1 {
			return nil, fmt.Errorf("row %d length mismatch", u+1)
		}
		matrix[u] = make([]float64, numMovies)
		for m := 0; m < numMovies; m++ {
			val, err := strconv.ParseFloat(row[m+1], 64)
			if err != nil {
				return nil, fmt.Errorf("parse float row %d col %d: %v", u+1, m+1, err)
			}
			matrix[u][m] = val
		}
	}

	return &MatrixData{
		Matrix:              matrix,
		MovieIndexToMovieID: movieIndexToMovieID,
		MovieIDToMovieIndex: movieIDToMovieIndex,
	}, nil
}
