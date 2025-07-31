package urand

import "golang.org/x/exp/rand"

var MaleNames = []string{
	"James", "John", "Robert", "Michael", "William",
	"David", "Richard", "Joseph", "Thomas", "Charles",
	"Christopher", "Daniel", "Matthew", "Anthony", "Donald",
	"Mark", "Paul", "Steven", "Andrew", "Kenneth",
	"George", "Joshua", "Kevin", "Brian", "Edward",
	"Ronald", "Timothy", "Jason", "Jeffrey", "Gary",
	"Ryan", "Nicholas", "Eric", "Steven", "Jacob",
	"Larry", "Frank", "Scott", "Justin", "Brandon",
	"Benjamin", "Samuel", "Patrick", "Alexander", "Gregory",
	"Ray", "Henry", "Alan", "Jerry", "Dennis",
}

var FemaleNames = []string{
	"Olivia", "Emma", "Ava", "Sophia", "Isabella",
	"Mia", "Amelia", "Harper", "Evelyn", "Abigail",
	"Ella", "Scarlett", "Grace", "Chloe", "Victoria",
	"Riley", "Zoey", "Nora", "Lily", "Lillian",
	"Addison", "Aria", "Avery", "Audrey", "Leah",
	"Layla", "Lillian", "Hannah", "Natalie", "Brooklyn",
	"Bella", "Zoe", "Mila", "Camila", "Aurora",
	"Lucy", "Stella", "Paisley", "Ellie", "Lila",
	"Caroline", "Maya", "Sophie", "Anna", "Aubrey",
	"Sadie", "Skylar", "Genesis", "Bella", "Claire",
}

func RandName() string {
	var ss = append(FemaleNames, MaleNames...)
	return ss[rand.Intn(len(ss))] + Digits(4)
}
