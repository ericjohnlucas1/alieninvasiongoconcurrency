# Alien Invasion Simulation

## Usage:

- The usage is as follows: `go run main.go <number_of_aliens>`
- The `main.go` file containing my implementation expects a file called `cities.txt` to be in the same directory it is run from. If no such file exists it will exit with an error. A sample `cities.txt` is provided
- At the top of the `main.go`, there is one integer variable for configuration: `maxmoves` the maximum number of moves, defaulted to 10,000

## Implementation:

- An undirected graph is the data structure used to represent the map. The adjacency list for the graph is represented as a map in Golang, with each city being mapped to set of all the cities there is a path to. This adjacency list is modified accordingly when a city is destroyed
- Aliens are unleashed by spawning goroutines, which repeat the procedure of roaming around the map until they encounter other aliens and die or reach the maximum number of moves allows
- There are two types of message events which are sent to the main goroutine through a channel. The main goroutine uses these messages to determine when all aliens are either dead or have reached the max number of moves so the program can gracefully exit. The message types are as follows:
        1) City destroyed message- Signals to the main thread that a city is destroyed and which aliens were killed in the process so the main thread can print the appropriate message.
        2) An alien is out of moves

- sync.Mutex is used to make sure shared data structures are not modified at the same time or read while being modified
- The aliens are not given “names” per say, but are given serial numbers. If there are N aliens, they will be assigned serial numbers {0,1,2,3,4,…,N-1}


## Assumptions:
- The instructions say that when 2 aliens end up in the same place they fight, kill, each other and destroy the city… I interpreted this as when 2 or more aliens end up in the same place they all kill each other…
- An alien deciding to remain in the same city is an allowed move. In fact, this is the only possible move when aliens become trapped because there are no roads out of the current city they are in
