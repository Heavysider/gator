package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/heavysider/gator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: cli login <userName>")
	}
	userName := cmd.Args[0]
	dbUser, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		return err
	}

	err = s.config.SetUser(dbUser.Name)
	if err != nil {
		return err
	}
	fmt.Printf("Current user was set to: %v\n", userName)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: cli register <userName>")
	}
	userName := cmd.Args[0]
	createionTime := time.Now()
	dbUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: createionTime,
		UpdatedAt: createionTime,
		Name:      userName,
	})
	if err != nil {
		return err
	}

	err = s.config.SetUser(dbUser.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User: %v was succesfully created!\n", userName)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error while resetting users: %v", err)
	}
	fmt.Println("users table was successfully reset")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	dbUsers, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range dbUsers {
		if user.Name == s.config.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: cli agg <duration between fetches(1s, 1m, 1h)>")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error parsing duration(%v) between fetches: %v", cmd.Args[0], err)
	}

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, currentUser database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("usage: cli addfeed <feedname> <feedurl>")
	}
	feedName := cmd.Args[0]
	feedURL := cmd.Args[1]
	createionTime := time.Now()

	dbFeed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: createionTime,
		UpdatedAt: createionTime,
		Name:      feedName,
		Url:       feedURL,
		UserID:    currentUser.ID,
	})
	if err != nil {
		return err
	}
	createionTime = time.Now()

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    currentUser.ID,
		FeedID:    dbFeed.ID,
		CreatedAt: createionTime,
		UpdatedAt: createionTime,
	})
	if err != nil {
		return err
	}

	fmt.Printf("You've followed the %v feed!\n", dbFeed.Name)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	dbFeeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	if len(dbFeeds) == 0 {
		fmt.Println("no feeds were added yet")
		return nil
	}

	var userUUIDs []uuid.UUID
	for _, dbFeed := range dbFeeds {
		userUUIDs = append(userUUIDs, dbFeed.UserID)
	}
	dbUsers, err := s.db.GetUsersByIds(context.Background(), userUUIDs)
	if err != nil {
		return err
	}
	userMap := map[uuid.UUID]database.User{}

	for _, dbUser := range dbUsers {
		userMap[dbUser.ID] = dbUser
	}

	for _, dbFeed := range dbFeeds {
		fmt.Printf("Feed name: %v, Feed URL: %v, User: %v\n", dbFeed.Name, dbFeed.Url, userMap[dbFeed.UserID].Name)
	}
	return nil
}

func handlerFollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: cli follow <url>")
	}

	dbFeed, err := s.db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}
	createionTime := time.Now()

	res, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    currentUser.ID,
		FeedID:    dbFeed.ID,
		CreatedAt: createionTime,
		UpdatedAt: createionTime,
	})
	if err != nil {
		return err
	}

	fmt.Printf("User: %v, just followed the %v feed!\n", res.UserName, res.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, currentUser database.User) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return err
	}

	for _, feedFollow := range feedFollows {
		fmt.Printf("%v\n", feedFollow.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: cli unfollow <url>")
	}

	dbFeed, err := s.db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	err = s.db.DeleteFeedFollowForUser(context.Background(), database.DeleteFeedFollowForUserParams{
		UserID: currentUser.ID,
		FeedID: dbFeed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("You've successfully unfollowed %v feed!", dbFeed.Name)
	return nil
}

func handlerBrowse(s *state, cmd command, currentUser database.User) error {
	var limit int
	var err error
	if len(cmd.Args) == 0 {
		limit = 2
	} else {
		limit, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("cli browse [limit(int)]")
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: currentUser.ID,
		Limit:  int32(limit),
	})

	for _, post := range posts {
		fmt.Println("Title:", post.Title)
		fmt.Println("Description:", post.Description)
		fmt.Println("Published at:", post.PublishedAt)
		fmt.Println("URL:", post.Url)
		fmt.Println("------------------------------------------------------------------------------")
	}
	return nil
}
