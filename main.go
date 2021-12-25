package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ReadLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func AppendToSection(section string, lines []string, item string) []string {
	for i, l := range lines {
		if l == section {
			for strings.TrimSpace(lines[i]) != "" {
				if lines[i] == item {
					return lines
				}
				i++
			}
			return append(append(lines[:i], item+"\r\n"), lines[i+1:]...)
		}
	}
	return append(append(lines, section), item)
}

func ReplaceBetweenSections(lines []string, start, end, content string) []string {
	startIndex := -1
	endIndex := -1
	for i, l := range lines {
		if l == start {
			startIndex = i
		}

		if l == end {
			endIndex = i
		}
	}
	if startIndex == -1 {
		panic("Start section not found: " + start)
	}
	if endIndex == -1 {
		panic("Start section not found: " + end)
	}
	return append(append(lines[:startIndex+2], content), lines[endIndex:]...)
}

func ToString(lines []string) string {
	result := ""
	for _, l := range lines {
		result += l + "\r\n"
	}
	return result
}

func WriteToFile(filename, s string) {
	err := os.WriteFile(filename, []byte(s), 0644)
	if err != nil {
		panic(err)
	}
}

func FindInstalledMaps(basepath string) []string {
	var matches []string
	err := filepath.Walk(basepath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match("*.kfm", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			fmt.Println("Found map: " + path)
			matches = append(matches, strings.Split(filepath.Base(path), ".")[0])
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return matches
}

func GenerateMapSection(mapnames []string) string {
	result := "\r\n"
	for _, m := range mapnames {
		result += fmt.Sprintf("[%s KFMapSummary]\r\n", m)
		result += fmt.Sprintf("MapName=%s\r\n\r\n", m)
	}
	return result
}

func ReplaceMapCycle(lines []string, customMaps []string) []string {
	defaultMaps := []string{"KF-Airship", "KF-AshwoodAsylum", "KF-Biolapse", "KF-Bioticslab", "KF-BlackForest", "KF-BurningParis", "KF-Catacombs", "KF-ContainmentStation", "KF-Desolation", "KF-DieSector", "KF-Dystopia2029", "KF-Moonbase", "KF-Elysium", "KF-EvacuationPoint", "KF-Farmhouse", "KF-HellmarkStation", "KF-HostileGrounds", "KF-InfernalRealm", "KF-KrampusLair", "KF-Lockdown", "KF-MonsterBall", "KF-Netherhold", "KF-Nightmare", "KF-Nuked", "KF-Outpost", "KF-PowerCore_Holdout", "KF-Prison", "KF-Sanitarium", "KF-Santasworkshop", "KF-ShoppingSpree", "KF-Spillway", "KF-SteamFortress", "KF-TheDescent", "KF-TragicKingdom", "KF-VolterManor", "KF-ZedLanding"}
	allMaps := append(defaultMaps, customMaps...)

	for i, l := range lines {
		if strings.HasPrefix(l, "GameMapCycles=") {

			mapString := ""
			for _, m := range allMaps {
				mapString += "\"" + m + "\","
			}

			lines[i] = fmt.Sprintf("GameMapCycles=(Maps=(%s))", mapString[:len(mapString)-1])
			return lines
		}
	}
	return lines
}

func main() {

	var steamItem []string
	if len(os.Args) < 2 {
		fmt.Println("No Cmdline arguments, reading from maps.txt")
		steamItem = ReadLines("maps.txt")
	} else {
		steamItem = os.Args[1:]
	}

	basePath, err := os.Getwd() //"C:\\Program Files (x86)\\Steam\\kf2_ds"
	if err != nil {
		panic(err)
	}

	engineConfigPath := basePath + "\\KFGame\\Config\\PCServer-KFEngine.ini"
	gameConfigPath := basePath + "\\KFGame\\Config\\PCServer-KFGame.ini"
	exePath := basePath + "\\Binaries\\Win64\\KFServer.exe"
	cachePath := basePath + "\\KFGame\\Cache"

	fmt.Println("Engine Config File:", engineConfigPath)
	fmt.Println("Game Config File:", gameConfigPath)

	engineConfig := ReadLines(engineConfigPath)
	gameConfig := ReadLines(gameConfigPath)

	for _, item := range steamItem {
		engineConfig = AppendToSection("[OnlineSubsystemSteamworks.KFWorkshopSteamworks]", ReadLines(engineConfigPath), "ServerSubscribedWorkshopItems="+item)
	}
	WriteToFile(engineConfigPath, ToString(engineConfig))
	fmt.Println("Updated:", engineConfigPath)
	fmt.Println("Starting KFServer.exe to fetch new maps, please close the KFServer windows once the maps are downloaded, find details in " + basePath + "\\Binaries\\Win64\\logs\\workshop_log.txt")

	cmd := exec.Command(exePath)
	cmd.Run()

	maps := FindInstalledMaps(cachePath)
	newGameConfig := ReplaceMapCycle(ReplaceBetweenSections(gameConfig, "[KF-Default KFMapSummary]", "[Boom KFWeeklyOutBreakInformation]", GenerateMapSection(maps)), maps)
	WriteToFile(gameConfigPath, ToString(newGameConfig))

	fmt.Println("PCServer-KFEngine.ini and PCServer-KFGame.ini updated successfully :)")
	fmt.Println("Following maps are installed: ")
	for _, m := range maps {
		fmt.Println("- ", m)
	}
}
