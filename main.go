package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aaronbolcerek/BlogAggregator/internal/config"
	"github.com/aaronbolcerek/BlogAggregator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	db *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title string `xml:"title"`
		Link string `xml:"link"`
		Description string `xml:"description"`
		Item []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title string `xml:"title"`
	Link string `xml:"link"`
	Description string `xml:"description"`
	PubDate string `xml:"pubDate"`
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.commands[cmd.name]
	if !ok {
		return fmt.Errorf("Error, command name not found")
	} 
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("No command argument\n")
	}
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		return fmt.Errorf("User does not exist\n")
	}
	err = s.config.SetUser(cmd.args[0]) 
	if err != nil {
		return err
	}
	fmt.Printf("User has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("No command argument\n")
	}
	ctx := context.Background()
	created_at := time.Now()
	updated_at := time.Now()
	user := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: created_at,
		UpdatedAt: updated_at,
		Name: cmd.args[0],
	}
	_, err := s.db.CreateUser(ctx, user)
	if err != nil {
		return err
	}
	err = s.config.SetUser(cmd.args[0]) 
	if err != nil {
		return err
	}
	fmt.Printf("User has been created\n")
	ctx = context.Background()
	user_details, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Print(user_details)
	return nil
}

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.ResetUsers(ctx)
	if err != nil {
		return fmt.Errorf("Error when resetting database\n")
	}
	return nil
}

func handlerUsers(s *state, cmd command) error {
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.config.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Incorrect number of arguments\n")
	}
	ctx := context.Background()
	current_user := s.config.CurrentUserName
	user, err := s.db.GetUser(ctx, current_user)
	if err != nil {
		return err
	}
	created_at := time.Now()
	updated_at := time.Now()
	feed_params := database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: created_at,
		UpdatedAt: updated_at,
		Name: cmd.args[0],
		Url: cmd.args[1],
		UserID: user.ID,
	}
	ctx = context.Background()
	feed, err := s.db.CreateFeed(ctx, feed_params)
	if err != nil {
		return err
	}
	ctx = context.Background()
	created_at = time.Now()
	updated_at = time.Now()
	feed_follow_params := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: created_at,
		UpdatedAt: updated_at,
		UserID: user.ID,
		FeedID: feed.ID,
	}
	_, err = s.db.CreateFeedFollow(ctx, feed_follow_params)
	if err != nil {
		return err
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}
	req.Header.Set("User-Agent", "gator")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}
	response_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RSSFeed{}, err
	}
	feed := RSSFeed{}
	err = xml.Unmarshal(response_body, &feed)
	if err != nil {
		return &RSSFeed{}, err
	}
	decoded_title := html.UnescapeString(feed.Channel.Title)
	decoded_description := html.UnescapeString(feed.Channel.Description)
	feed.Channel.Title = decoded_title
	feed.Channel.Description = decoded_description
	for index, _ := range feed.Channel.Item {
		decoded_title = html.UnescapeString(feed.Channel.Item[index].Title)
		decoded_description = html.UnescapeString(feed.Channel.Item[index].Description)
		feed.Channel.Item[index].Title = decoded_title
		feed.Channel.Item[index].Description = decoded_description
	}
	return &feed, nil
}

func handlerFeeds(s *state, cmd command) error {
	ctx := context.Background()
	feed, err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}
	for i := range feed {
		fmt.Println(feed[i].Name)
		fmt.Println(feed[i].Url)
		fmt.Println(feed[i].UserName)
	}
	return nil
}

func handlerFollow(s *state, cmd command) error {
	ctx := context.Background()
	user, err := s.db.GetUser(ctx, s.config.CurrentUserName)
	if err != nil {
		return err
	}
	ctx = context.Background()
	feed, err := s.db.GetFeed(ctx, cmd.args[0])
	if err != nil {
		return err
	}
	ctx = context.Background()
	created_at := time.Now()
	updated_at := time.Now()
	feed_follow_params := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: created_at,
		UpdatedAt: updated_at,
		UserID: user.ID,
		FeedID: feed.ID,
	}
	_, err = s.db.CreateFeedFollow(ctx, feed_follow_params)
	if err != nil {
		return err
	}
	return nil
}

func handlerFollowing(s *state, cmd command) error {
	ctx := context.Background()
	user, err := s.db.GetUser(ctx, s.config.CurrentUserName)
	if err != nil {
		return err
	}
	ctx = context.Background()
	follow_feeds, err := s.db.GetFollowing(ctx, user.ID)
	if err != nil {
		return err
	}
	for i := range follow_feeds {
		fmt.Println(follow_feeds[i].FeedName)
	}
	return nil
}

func main() {
	original_config, err := config.Read()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", original_config.DbUrl)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	user_state := state{
		db: dbQueries,
		config: &original_config,
	}
	cmds := commands{
		commands: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", handlerFollow)
	cmds.register("following", handlerFollowing)
	arguments := os.Args
	if len(arguments) < 2 {
		fmt.Printf("Please provide more than 1 argument\n")
		os.Exit(1)
	}
	cmd := command{name: arguments[1], args: arguments[2:]}
	err = cmds.run(&user_state, cmd)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}