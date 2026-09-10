package types

// ForumBoard is one board under a forum category.
type ForumBoard struct {
	BoardID     int    `json:"board_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ForumPost is one post in a forum topic.
type ForumPost struct {
	ID        int    `json:"id"`
	Number    int    `json:"number"`
	CreatedAt string `json:"created_at"`
	Body      string `json:"body"`
	User      struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}

// ForumTopic is one topic from /forum/topics.
type ForumTopic struct {
	ID                int    `json:"id"`
	Title             string `json:"title"`
	CreatedAt         string `json:"created_at"`
	LastPostCreatedAt string `json:"last_post_created_at"`
	PostsCount        int    `json:"posts_count"`
	User              struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}
