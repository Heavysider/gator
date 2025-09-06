package main

import (
	"context"
	"fmt"
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
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Println(*feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("usage: cli addfeed <feedname> <feedurl>")
	}
	feedName := cmd.Args[0]
	feedURL := cmd.Args[1]
	createionTime := time.Now()
	currentUser, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

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

	fmt.Println(dbFeed)
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
