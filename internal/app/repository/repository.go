package repository

import "fmt"

var Repository = make(map[int]string)

func InitRepository() {
	Repository = make(map[int]string)
}

func getNextID() int {
	return len(Repository) + 1
}

func SaveURL(URL string) int {
	id := getNextID()
	Repository[id] = URL
	return id
}

func GetURL(id int) (string, error) {
	URL, ok := Repository[id]
	if !ok {
		return "", fmt.Errorf("not found")
	}
	return URL, nil
}
