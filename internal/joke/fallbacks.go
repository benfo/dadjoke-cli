package joke

// FallbackJokes provides offline dad jokes when the network is unavailable and no cache exists.
var FallbackJokes = []string{
	"Why do fathers take an extra pair of socks golfing? In case they get a hole in one.",
	"I told my doctor that I broke my arm in two places. He told me to stop going to those places.",
	"What do you call a fake noodle? An impasta.",
	"How does a penguin build its house? Igloos it together.",
	"Why don't skeletons fight each other? They don't have the guts.",
	"What do you call cheese that isn't yours? Nacho cheese.",
	"Why did the scarecrow win an award? Because he was outstanding in his field.",
	"I'm reading a book on anti-gravity. It's impossible to put down.",
	"Why do melons have weddings? Because they cantaloupe.",
	"What did the ocean say to the shore? Nothing, it just waved.",
}

// GetFallbackJoke returns a deterministic fallback joke for a given seed (such as YYYY-MM-DD).
func GetFallbackJoke(seed string) string {
	var hash uint32
	for i := 0; i < len(seed); i++ {
		hash = hash*31 + uint32(seed[i])
	}
	index := int(hash % uint32(len(FallbackJokes)))
	return FallbackJokes[index]
}
