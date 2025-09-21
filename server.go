package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var YOUTUBE_API_KEY string

type DATAS struct {
	IMAGE         string `json:"image"`
	TITLE         string `json:"title"`
	URL           string `json:"url"`
	CHANNEL       string `json:"channel"`
	VIEW_COUNT    string `json:"view_count"`
	LIKE_COUNT    string `json:"like_count"`
	COMMENT_COUNT string `json:"comment_count"`
}

type YouTubeResponse struct {
	Items []struct {
		ID         string     `json:"id"`
		Snippet    Snippet    `json:"snippet"`
		Statistics Statistics `json:"statistics"`
	} `json:"items"`
}

type Snippet struct {
	Title        string               `json:"title"`
	ChannelTitle string               `json:"channelTitle"`
	ChannelID    string               `json:"channelId"`
	Thumbnails   map[string]Thumbnail `json:"thumbnails"`
}

type Thumbnail struct {
	URL string `json:"url"`
}

type Statistics struct {
	ViewCount    string `json:"viewCount"`
	LikeCount    string `json:"likeCount"`
	CommentCount string `json:"commentCount"`
}

type KeyWordRequest struct {
	Keyword string `json:"keyword"`
}

type SearchResponse struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	RegionCode    string `json:"regionCode,omitempty"`
	PageInfo      struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []struct {
		Kind string `json:"kind"`
		Etag string `json:"etag"`
		ID   struct {
			Kind       string `json:"kind"`
			VideoID    string `json:"videoId"`
			PlaylistID string `json:"playlistId,omitempty"`
			ChannelID  string `json:"channelId,omitempty"`
		} `json:"id"`
		Snippet struct {
			PublishedAt  string `json:"publishedAt"`
			ChannelID    string `json:"channelId"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			ChannelTitle string `json:"channelTitle"`
			Thumbnails   struct {
				Default struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"default"`
				Medium struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"medium"`
				High struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"high"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}

type YouTubeVideosResponse struct {
	Kind  string `json:"kind"`
	Etag  string `json:"etag"`
	Items []struct {
		Kind    string `json:"kind"`
		Etag    string `json:"etag"`
		ID      string `json:"id"`
		Snippet struct {
			PublishedAt  string `json:"publishedAt"`
			ChannelID    string `json:"channelId"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			ChannelTitle string `json:"channelTitle"`
			Thumbnails   struct {
				Default struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"default"`
				Medium struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"medium"`
				High struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"high"`
				Maxres struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"maxres,omitempty"`
			} `json:"thumbnails"`
		} `json:"snippet"`
		Statistics struct {
			ViewCount    string `json:"viewCount"`
			LikeCount    string `json:"likeCount"`
			CommentCount string `json:"commentCount"`
		} `json:"statistics"`
	} `json:"items"`
	PageInfo struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
}

var res_texts = []DATAS{}

func youtube_main_data_text(c *gin.Context) {
	res_texts = []DATAS{}

	youtube_trend_data_get()
	c.JSON(http.StatusOK, res_texts)
}

func youtube_trend_data_get() {

	if YOUTUBE_API_KEY == "" {
		fmt.Println("YouTube API Key is not set")
	}

	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&chart=mostPopular&maxResults=48&regionCode=JP&key=%s", YOUTUBE_API_KEY)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	var result YouTubeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
	}

	res_texts = []DATAS{}

	for _, item := range result.Items {
		imageURL := ""
		if thumb, ok := item.Snippet.Thumbnails["maxres"]; ok {
			imageURL = thumb.URL
		} else if thumb, ok := item.Snippet.Thumbnails["high"]; ok {
			imageURL = thumb.URL
		} else if thumb, ok := item.Snippet.Thumbnails["medium"]; ok {
			imageURL = thumb.URL
		} else if thumb, ok := item.Snippet.Thumbnails["default"]; ok {
			imageURL = thumb.URL
		}

		res_texts = append(res_texts, DATAS{
			IMAGE:         imageURL,
			TITLE:         item.Snippet.Title,
			URL:           fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID),
			CHANNEL:       fmt.Sprintf("https://www.youtube.com/channel/%s", item.Snippet.ChannelID),
			VIEW_COUNT:    item.Statistics.ViewCount,
			LIKE_COUNT:    item.Statistics.LikeCount,
			COMMENT_COUNT: item.Statistics.CommentCount,
		})
	}
}

func youtube_key_word_data_text(c *gin.Context) {
	var req KeyWordRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	result := keyword_trend_get(req.Keyword)
	c.JSON(http.StatusOK, result)
}

func keyword_trend_get(keyword string) []DATAS {
	if YOUTUBE_API_KEY == "" {
		fmt.Println("YouTube API Key is not set")
	}

	encoding_keyword := url.QueryEscape(keyword)
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&maxResults=48&q=%s&type=video&key=%s", encoding_keyword, YOUTUBE_API_KEY)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
	}

	video_ids := ""
	for i, item := range result.Items {
		if i > 0 {
			video_ids += ","
		}
		video_ids += item.ID.VideoID
	}

	if video_ids == "" {
		return []DATAS{}
	}

	url = fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=%s&key=%s", video_ids, YOUTUBE_API_KEY)

	vresp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
	}
	defer vresp.Body.Close()

	vbody, err := ioutil.ReadAll(vresp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	var resultVideos YouTubeVideosResponse
	if err := json.Unmarshal(vbody, &resultVideos); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
	}

	res := []DATAS{}
	for _, item := range resultVideos.Items {
		imageURL := ""
		if item.Snippet.Thumbnails.Maxres.URL != "" {
			imageURL = item.Snippet.Thumbnails.Maxres.URL
		} else if item.Snippet.Thumbnails.High.URL != "" {
			imageURL = item.Snippet.Thumbnails.High.URL
		} else if item.Snippet.Thumbnails.Medium.URL != "" {
			imageURL = item.Snippet.Thumbnails.Medium.URL
		} else if item.Snippet.Thumbnails.Default.URL != "" {
			imageURL = item.Snippet.Thumbnails.Default.URL
		}

		res = append(res, DATAS{
			IMAGE:         imageURL,
			TITLE:         item.Snippet.Title,
			URL:           fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID),
			CHANNEL:       fmt.Sprintf("https://www.youtube.com/channel/%s", item.Snippet.ChannelID),
			VIEW_COUNT:    item.Statistics.ViewCount,
			LIKE_COUNT:    item.Statistics.LikeCount,
			COMMENT_COUNT: item.Statistics.CommentCount,
		})
	}
	return res
}

func main() {
	YOUTUBE_API_KEY = os.Getenv("YOUTUBE_API_KEY")
	if YOUTUBE_API_KEY == "" {
		fmt.Println("YOUTUBE_API_KEY is not set in environment variables")
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"https://youtube-trends-finder.vercel.app"}, // フロントのURL
    AllowMethods:     []string{"GET", "POST", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
    }))

	// router.Static("/", "./frontend")
	router.GET("/youtube/main_trend/data", youtube_main_data_text)
	router.POST("/youtube/key_word_trend/data", youtube_key_word_data_text)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run("0.0.0.0:" + port)
}




