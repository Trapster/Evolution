package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

//Items to use
type items struct {
	weights []float64
}

// discription of an individual
type genome struct {
	gen       []bool
	sumValues float64
	fit       float64
}

// create a new individual with a fresh random genome
func newGene(l int) genome {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	s := make([]bool, l)
	for i := range s {
		s[i] = r.Float64() < 0.5
	}

	return genome{gen: s}
}

func newStuff(l int) items {

	// Uncomment to make a bag filled with all the integers up to starting from 1...
	s := make([]float64, l)
	r := 1.0
	for i := range s {
		s[i] = r
		r++
	}

	// Uncomment to allow for random float64 numbers to be the items to choose from, this is more realistic
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// s := make([]float64, l)
	// for i := range s {
	// 	s[i] = r.Float64() * 10
	// }

	return items{s}
}

//Fitness calculator
func totalValue(g genome, stuff []float64) float64 {

	var sumValues float64

	for i := range g.gen {
		if g.gen[i] {
			sumValues = sumValues + stuff[i]
		}
	}
	return sumValues
}

// This is the fitness function. Here I am using a simple quadratic.
func valueToFit(v float64, t float64) float64 {

	f := -((t - v) * (t - v)) + 1

	return f
}

//input pupulation output slice of same size with a fitness assignment
func assignFitness(p []genome, stuff []float64, target float64) []genome {

	// var resulting []genome

	for i := range p {
		go func(p []genome, stuff []float64, target float64, i int) {
			value := totalValue(p[i], stuff)
			p[i].sumValues = value
			fit := valueToFit(value, target)
			p[i].fit = fit
		}(p, stuff, target, i)

	}

	return p
}

// take in population and fitness and return the selected
func createSelection(p []genome, e int, s int) ([]genome, []genome) {

	elite := make([]genome, e)
	for i := 0; i < e; i++ {
		elite[i].gen = make([]bool, len(p[i].gen))
		copy(elite[i].gen, p[i].gen)
		elite[i].fit = p[i].fit
	}

	survivors := make([]genome, s)
	for i := 0; i < s; i++ {
		survivors[i].gen = make([]bool, len(p[i].gen))
		copy(survivors[i].gen, p[i].gen)
		survivors[i].fit = p[i].fit
	}

	return elite, survivors

}

func makeBabies(a genome, b genome, geneSize int) []genome {
	//babies per couple
	bpc := 7
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var newBabies []genome
	for p := 0; p < bpc; p++ {
		var babyGene genome
		for i := 0; i < geneSize; i++ {
			if r.Float64() < 0.5 {
				babyGene.gen = append(babyGene.gen, a.gen[i])
			} else {
				babyGene.gen = append(babyGene.gen, b.gen[i])
			}

		}
		newBabies = append(newBabies, babyGene)
	}
	return newBabies
}

// // take in selected and create crossover
func crossover(s []genome, gSize int) []genome {
	//split parents
	var crossedGenomes []genome

	ch := make(chan []genome)
	for i := 0; i < len(s)-1; i++ {
		go func(i int, gSize int) {
			ch <- makeBabies(s[i], s[i+1], gSize)
		}(i, gSize)
	}
	for i := 0; i < len(s)-1; i++ {
		j := <-ch
		crossedGenomes = append(crossedGenomes, j...)
	}
	return crossedGenomes
}

// // randomly mutate
func mutation(p []genome, mutationProbability float64) []genome {

	for i := range p {

		go func(i int, mutationProbability float64) {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			if r.Float64() < mutationProbability {
				randomIndex := rand.Intn(len(p[i].gen))
				if p[i].gen[randomIndex] {
					p[i].gen[randomIndex] = false
					p[i].fit = 0

				} else {
					p[i].gen[randomIndex] = true
					p[i].fit = 0
				}
			}
		}(i, mutationProbability)
	}
	return p
}

func main() {

	//setup
	noIndividuals := 2000
	epochs := 100
	geneSize := 237
	target := 821
	noElite := 5
	mutationProbability := 0.8
	previousBest := -1.0 * 10.0 * 100000.0

	// items that can fit in the bag
	stuff := newStuff(geneSize)

	// randomly initialise the starting population.
	population := make([]genome, (noIndividuals))
	for i := range population {
		population[i] = newGene(geneSize)
		// fmt.Println(population[i])
	}

	// assign fitness
	population = assignFitness(population, stuff.weights, float64(target))
	// loop over epochs
	for i := 0; i < epochs; i++ {

		// select Elite and Survivor (dont forget to DEEP copy )
		elite, survivors := createSelection(population, noElite, noIndividuals)

		// randomly initialise outsiders to the population. These act as wild cards
		outsiders := make([]genome, 10)
		for i := range outsiders {
			outsiders[i] = newGene(geneSize)
		}

		// Introduce the outsiders to the survivors from the previous generation
		population = append(survivors, outsiders...)

		//Shuffle the population to achieve random mixing
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(population), func(i, j int) { population[i], population[j] = population[j], population[i] })

		// crossover - otherwise known as mating...
		population = crossover(population, geneSize)

		// mutate
		population = mutation(population, mutationProbability)

		//elite
		population = append(elite, population...)

		// assign fitness
		population = assignFitness(population, stuff.weights, float64(target))

		//sort desc just in case elite have been replaced
		sort.SliceStable(population, func(i, j int) bool { return population[i].fit > population[j].fit })

		//If there has been an exact match, stop looping.
		if population[0].fit == 1 {
			break
		} else { // Prints any improvements made in the past generation
			if population[0].fit > previousBest {
				fmt.Println("Epoch:", i, " Best so far, value (target = ", target, "): ", population[0].sumValues, " Fit (target = 1): ", population[0].fit)
				previousBest = population[0].fit
			}

		}

	}
	if population[0].fit == 1 {
		fmt.Println("Found exact match")
	} else {
		fmt.Println("Best match")
	}
	fmt.Println("gene: ", population[0].gen)
	fmt.Println("fit: ", population[0].fit)
}
