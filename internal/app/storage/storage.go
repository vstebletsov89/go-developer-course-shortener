package storage

var Repository = make(map[int]string)

func getNextID() int {
	return len(Repository) + 1
}

func SaveURL(URL string) int {
	id := getNextID()
	Repository[id] = URL
	return id
}

func GetURL(id int) string {
	URL, ok := Repository[id]
	if !ok {
		return ""
	}
	return URL
}
