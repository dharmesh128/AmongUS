package main

import (
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	randomtext "AmongUS/dataBreach" // Import package for generating random text
	redis2 "AmongUS/redis"          // Import Redis package
)

// Player represents a player in the game
type Player struct {
	Name       string   // Name of the player
	IsImpostor bool     // Indicates whether the player is an impostor or not
	Metadata   Metadata // Metadata for player
}

// Task represents a task in the game
type Task struct {
	Name       string   // Name of the task
	IsDone     bool     // Indicates whether the task is completed or not
	Metadata   Metadata // Metadata for task
	Taxonomies []string // Taxonomies for task
}

// Metadata represents metadata associated with players and tasks
type Metadata struct {
	CreatedAt time.Time // Creation time
	UpdatedAt time.Time // Last update time
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator with current time

	rdb := redis2.NewClient() // Create a new Redis client

	// Check if the connection to Redis is successful
	pong, err := rdb.Ping(redis2.Ctx).Result()
	if err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return
	}
	fmt.Println("Connected to Redis:", pong)

	// Read players' data from JSON file
	players, err := readPlayers("players.json")
	if err != nil {
		fmt.Println("Error reading players:", err)
		return
	}

	// Read tasks' data from JSON file
	tasks, err := readTasks("tasks.json")
	if err != nil {
		fmt.Println("Error reading tasks:", err)
		return
	}

	// Randomly assign one player as an impostor
	assignRandomImpostor(players)

	// Check if player names are unique
	if !areNamesUnique(players) {
		fmt.Println("Error: Player names are not unique")
		return
	}

	// Check if there is at least one impostor among players
	if !hasImpostor(players) {
		fmt.Println("Error: At least one player must be an impostor")
		return
	}

	// Print players after impostor assignment
	fmt.Println("\nPlayers after impostor assignment:")
	for _, player := range players {
		fmt.Printf("Name: %s, IsImpostor: %t\n", player.Name, player.IsImpostor)
	}

	var wg sync.WaitGroup

	// Launch goroutine for each player to complete a random task
	for _, player := range players {
		wg.Add(1)
		go func(p *Player) {
			defer wg.Done()
			p.completeTask(rdb, tasks[rand.Intn(len(tasks))])
		}(player)
	}

	wg.Wait() // Wait for all tasks to complete

	time.Sleep(2 * time.Second) // Introduce a delay for task completion before calling a meeting

	callMeeting(rdb, players) // Call a meeting after tasks are completed
}

// completeTask simulates a player completing a task
func (p *Player) completeTask(rdb *redis.Client, task *Task) {
	if p.IsImpostor {
		fmt.Printf("%s sabotages task: %s\n", p.Name, task.Name)
	} else {
		fmt.Printf("%s completes task: %s\n", p.Name, task.Name)
		task.IsDone = true
		task.Metadata.UpdatedAt = time.Now() // Update task metadata
		if err := updateTask(rdb, task); err != nil {
			fmt.Println("Error updating task:", err)
		}
	}
}

// callMeeting simulates calling a meeting in the game
func callMeeting(rdb *redis.Client, players []*Player) {
	fmt.Println("Emergency meeting called!")
	impostorText := randomtext.Encrypt("SENSITIVE INFORMATION", "EvilAlwaysWins") // Generate random text for impostor
	votedPlayer, isImpostor := getUserVote(players)
	if votedPlayer != nil {
		fmt.Printf("Player %s was voted out!\n", votedPlayer.Name)
		if isImpostor {
			if err := checkImpostor(rdb, votedPlayer, impostorText); err != nil {
				fmt.Println("Error checking impostor:", err)
				return
			}
		} else {
			if votedPlayer.IsImpostor {
				fmt.Println("Impostor Caught! Crewmates win!")
				fmt.Println("Impostor message:", impostorText)
				fmt.Println("DECRYPTING BEEP BOOP BEEP")
				decryptedText := randomtext.Decrypt(impostorText, "EvilAlwaysWins")
				fmt.Println("Impostor message:", decryptedText)
			} else {
				fmt.Println("Innocent Ejected 😔 Impostors win!")
				fmt.Println("Impostor message:", impostorText)
			}
		}
	} else {
		fmt.Println("Vote failed! No one was ejected.")
	}
}

// getUserVote simulates getting a player's vote during a meeting
func getUserVote(players []*Player) (*Player, bool) {
	fmt.Println("Voting time! Choose a player to eject:")

	for i, player := range players {
		fmt.Printf("%d. %s\n", i+1, player.Name)
	}

	var voteIndex int
	fmt.Print("Enter the number of the player you want to vote out: ")
	fmt.Scanln(&voteIndex)

	if voteIndex < 1 || voteIndex > len(players) {
		fmt.Println("Invalid vote!")
		return nil, false
	}

	return players[voteIndex-1], players[voteIndex-1].IsImpostor
}

// updateTask updates the status of a task in Redis
func updateTask(rdb *redis.Client, task *Task) error {
	return rdb.HSet(redis2.Ctx, "tasks", task.Name, task.IsDone).Err()
}

// checkImpostor checks if all impostors have been voted out
func checkImpostor(rdb *redis.Client, ejectedPlayer *Player, impostorText string) error {
	players, err := getPlayers(rdb)
	if err != nil {
		return err
	}

	impostorsLeft := 0
	for _, player := range players {
		if player.IsImpostor {
			impostorsLeft++
		}
	}

	if impostorsLeft > 0 {
		fmt.Println("Innocent Ejected 😔 Impostors win!")
		fmt.Println("Impostor message:", impostorText)
	} else {
		fmt.Println("Imposter Caught! Crewmates win!")
		fmt.Println("Impostor message:", impostorText)
		fmt.Println("DECRYPTING BEEP BOOP BEEP")
		decryptedText := randomtext.Decrypt(impostorText, "EvilAlwaysWins")
		fmt.Println("Impostor message:", decryptedText)
	}

	return nil
}

// getPlayers retrieves player data from Redis
func getPlayers(rdb *redis.Client) ([]*Player, error) {
	res, err := rdb.HGetAll(redis2.Ctx, "players").Result()
	if err != nil {
		return nil, err
	}

	players := make([]*Player, len(res))
	i := 0
	for name, isImpostor := range res {
		isImpostorBool, _ := strconv.ParseBool(isImpostor)
		players[i] = &Player{Name: name, IsImpostor: isImpostorBool}
		i++
	}

	return players, nil
}

// readPlayers reads player data from a JSON file
func readPlayers(filename string) ([]*Player, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var players []*Player
	if err := json.NewDecoder(file).Decode(&players); err != nil {
		return nil, err
	}

	return players, nil
}

// readTasks reads task data from a JSON file
func readTasks(filename string) ([]*Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []*Task
	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// assignRandomImpostor randomly assigns one player as an impostor
func assignRandomImpostor(players []*Player) {
	for _, player := range players {
		player.IsImpostor = rand.Intn(2) == 0 // Randomly assign impostor status
	}
}

// areNamesUnique checks if player names are unique
func areNamesUnique(players []*Player) bool {
	nameMap := make(map[string]bool)
	for _, player := range players {
		if nameMap[player.Name] {
			return false
		}
		nameMap[player.Name] = true
	}
	return true
}

// hasImpostor checks if there is at least one impostor among players
func hasImpostor(players []*Player) bool {
	for _, player := range players {
		if player.IsImpostor {
			return true
		}
	}
	return false
}
