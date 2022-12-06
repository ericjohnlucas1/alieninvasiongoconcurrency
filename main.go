package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
)

//configurable parameter for maximum moves allowed per alien
var maxmoves int = 10000

//a struct I defined for debug purposes to print when an alien transitions from one city to another
type citytransition struct {
	alien   int
	oldcity string
	newcity string
}

//a struct to signal through a channel to the main thread that a city was destroyed and which cities were destroyed with it
type citydestroyed struct {
	alienserials []int
	cityname     string
}

//a function defining the routine of an unleashed alien
func unleashedalien(alienserial int, startcity string, messages chan interface{}, citygraph map[string]map[string]bool, cities []string, alienposition map[string]map[int]bool, aliensalive map[int]bool, lock *sync.Mutex) {
	moves := 0
	for {
		//time.Sleep(1 * time.Second)
		lock.Lock()
		if _, ok := aliensalive[alienserial]; !ok { //alien is dead(release lock and exit goroutine)
			lock.Unlock()
			break
		}

		//an alien moves into a city at random, I add '+1' to the move options because I count the decision to stay in the current city as a move
		nextinx := rand.Intn(len(citygraph[startcity]) + 1)
		newcity := startcity
		if nextinx < len(citygraph[startcity]) {
			possibledestinationcities := []string{}
			for possibledestinationcity := range citygraph[startcity] {
				possibledestinationcities = append(possibledestinationcities, possibledestinationcity)
			}
			newcity = possibledestinationcities[nextinx]
		}

		nextmove := citytransition{alien: alienserial, oldcity: startcity, newcity: newcity}

		//alien leaves start city
		if ok, _ := alienposition[startcity][alienserial]; ok {
			delete(alienposition[startcity], alienserial)
		}
		messages <- nextmove

		//if the alien moves into a city where there are already other aliens, all the aliens in this city die and the city(and all the roads to the city get destroyed)
		if len(alienposition[newcity]) > 0 {
			alienskilled := []int{}
			for k := range alienposition[newcity] {
				alienskilled = append(alienskilled, k)
			}
			alienskilled = append(alienskilled, alienserial)
			for _, a := range alienskilled {
				delete(aliensalive, a)
			}
			delete(alienposition, newcity) //remove all aliens from the destroyed city
			for outcity := range citygraph[newcity] {
				delete(citygraph[outcity], newcity) //delete the roads leading back into this city
			}
			delete(citygraph, newcity) //delete the destroyed city and all roads in from the graph
			message := citydestroyed{cityname: newcity, alienserials: alienskilled}
			messages <- message
			lock.Unlock()
			break // alien dies (exit goroutine)
		}
		alienposition[newcity][alienserial] = true //update alien position with new destination city
		lock.Unlock()

		//increment the moves counter.. if the maximum number of allowed moves are reached, signal that to the main goroutine and terminate this goroutine
		moves++
		if moves == maxmoves {
			messages <- alienserial
			break
		}

		startcity = newcity
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Exactly one command line argument must be specified, indicating the number of aliens")
		os.Exit(1)
	}
	numaliens, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	citygraph := map[string]map[string]bool{}  //a adjacency list representing a graph data structure to keep track of connectivity between cities. This graph will be modified accordingly when a city is destroyed.
	cities := []string{}                       //a list of all the initial cities
	alienposition := map[string]map[int]bool{} //a real time indicator of which aliens are in each city
	events := make(chan interface{})           //indicates when a city is destroyed or an alien has reached max moves
	aliensalive := map[int]bool{}              //keeps track of aliens currently alive
	lock := &sync.Mutex{}                      //mutex lock to ensure only one goroutine can modify shared data structures at once

	dat, err := os.ReadFile("cities.txt")
	if err != nil {
		fmt.Println("Error reading input file 'cities.txt'")
		os.Exit(1)
	}

	//iterate through input lines and load the edges into the graph adjacency list
	for _, line := range strings.Split(string(dat), "\n") {
		if len(strings.TrimSpace(line)) > 0 {
			fields := strings.Fields(line)
			alienposition[fields[0]] = map[int]bool{} //just initializing this map for each city
			for i, field := range fields {
				if i > 0 {
					if citygraph[fields[0]] == nil {
						citygraph[fields[0]] = map[string]bool{}
					}
					edge := strings.Split(field, "=")
					if len(edge) != 2 {
						fmt.Println("Error parsing input file. Edge assignments must contain exactly one '='")
					}
					citygraph[fields[0]][edge[1]] = true
					if citygraph[edge[1]] == nil {
						citygraph[edge[1]] = map[string]bool{}
					}
					citygraph[edge[1]][fields[0]] = true
					alienposition[edge[1]] = map[int]bool{} //just initializing this map for each city
				}
			}
		}
	}

	for k, _ := range alienposition {
		cities = append(cities, k) //just building the list of all the cities
	}

	for i := 0; i < numaliens; i++ {
		aliensalive[i] = true //building the map to keep track of all aliens currently alive in real time
	}

	//below we unleash each of the aliens into a random city. the lock is acquired to ensure that aliens cannot begin roaming until they are all unleashed
	lock.Lock()
	for i := 0; i < numaliens; i++ {
		cityindex := rand.Intn(len(cities))
		alienposition[cities[cityindex]][i] = true
		go unleashedalien(i, cities[cityindex], events, citygraph, cities, alienposition, aliensalive, lock)
	}
	lock.Unlock()

	//keep recieving events that aliens are dead or aliens are out of moves until aliens no longer remain
	for event := range events {

		//I commented out this block of code for submission of this assignment. I had used these debug statements to print when an alien transitions cities
		/*
			t, ok := event.(citytransition)
			if ok {
				fmt.Println(t)
			}*/

		//this event indicates that a city was destroyed and aliens were killed
		destruction, ok := event.(citydestroyed)
		if ok {
			fmt.Printf("%v has been destroyed by aliens with the following serial numbers: %v\n", destruction.cityname, strings.Join(strings.Fields(fmt.Sprint(destruction.alienserials)), ", "))
			numaliens -= len(destruction.alienserials)
			if numaliens == 0 {
				break
			}
		}

		//this event indicates that an alien is out of moves
		_, ok = event.(int)
		if ok {
			numaliens--
			if numaliens == 0 {
				break
			}
		}
	}
}
