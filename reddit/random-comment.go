package main

/* This program retrieves the text of a single random comment from reddit and displays it with no context whatsoever. */

import "net/http"
import "encoding/json"
import "math/rand"
import "time"
import "fmt"
import "os"

const UserAgent = "random comment grabber v1.0 by /u/BranLwyd" /* UserAgent to send along with requests. */
const RandomURL = "http://reddit.com/r/all/random.json"        /* URL to request. Should be a random.json API endpoint. */
const SecondsPerRequest = 2                                    /* Number of seconds between network requests. Per Reddit API rules, this should never be set to less than 2. */

var requestThrottler <-chan time.Time
var httpClient http.Client = http.Client{}

func RequestRandom() (interface{}, error) {
	if requestThrottler == nil {
		/* on the first call to RequestRandom, don't wait */
		requestThrottler = time.Tick(SecondsPerRequest * time.Second)
	} else {
		<-requestThrottler
	}

	req, err := http.NewRequest("GET", RandomURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request for %s got status %s", RandomURL, resp.Status)
	}

	var decodedJson interface{}
	jsonDecoder := json.NewDecoder(resp.Body)
	err = jsonDecoder.Decode(&decodedJson)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return decodedJson, nil
}

func parseListing(data map[string]interface{}, comments chan string, children chan int) {
	childrenIntf, ok := data["children"]
	if ok {
		children <- 1
		go parseJsonPiece(childrenIntf, comments, children)
	}
}

func parseComment(data map[string]interface{}, comments chan string, children chan int) {
	/* get body & pass to master */
	bodyIntf, ok := data["body"]
	if ok {
		body, ok := bodyIntf.(string)
		if ok {
			comments <- body
		}
	}

	/* parse replies */
	repliesIntf, ok := data["replies"]
	if ok {
		children <- 1
		go parseJsonPiece(repliesIntf, comments, children)
	}
}

func parseJsonPiece(jsonIntf interface{}, comments chan string, children chan int) {
	defer func() { children <- -1 }()

	switch jj := jsonIntf.(type) {
	case []interface{}:
		/* handle arrays by recursively handling each subpiece */
		for _, pieceIntf := range jj {
			children <- 1
			go parseJsonPiece(pieceIntf, comments, children)
		}

	case map[string]interface{}:
		kindIntf, ok := jj["kind"]
		if !ok {
			/* every thing is supposed to have a kind... */
			return
		}
		kind, ok := kindIntf.(string)
		if !ok {
			/* kind should always be a string */
			return
		}

		dataIntf, ok := jj["data"]
		if !ok {
			return
		}
		data, ok := dataIntf.(map[string]interface{})
		if !ok {
			return
		}

		switch kind {
		case "Listing":
			parseListing(data, comments, children)

		case "t1":
			parseComment(data, comments, children)
		}
	}
}

func ParseJson(json interface{}) []string {
	commChan := make(chan string)
	childChan := make(chan int)
	comments := make([]string, 0, 16)

	/* kick off parsing the json */
	go parseJsonPiece(json, commChan, childChan)

	/* wait for children to finish & collate results */
	children := 1
	for children > 0 {
		select {
		case comment := <-commChan:
			n := len(comments)
			if n == cap(comments) {
				newComments := make([]string, n, n<<1)
				for i, v := range comments {
					newComments[i] = v
				}
				comments = newComments
			}
			comments = comments[:n+1]
			comments[n] = comment

		case childCount := <-childChan:
			children += childCount
		}
	}

	return comments
}

func RandomComment() (string, error) {
	for {
		commentJson, err := RequestRandom()
		if err != nil {
			return "", err
		}

		comments := ParseJson(commentJson)
		if len(comments) == 0 {
			/* we got a post with no comments. try again. */
			continue
		}

		index := rand.Intn(len(comments))
		return comments[index], nil
	}

	panic("unreachable")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	comment, err := RandomComment()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(comment)
}
