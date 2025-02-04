/*
   Todoro22/Gobuster-projet on Git

   Rappel : Ceci un un programme à titre éducatif et de démonstration.

   Description :
     Ce programme implémente un scan de type Gobuster en Go. Son rôle est de découvrir
     des chemins/fichiers potentiellement cachés sur un serveur web, en testant différents
     mots-clés ou patterns définis dans un dictionnaire (ex: "git_wordlist.txt").
     Référence : https://github.com/fatih/color (pour la coloration des sorties).
*/

package main

// Importation des packages nécessaires pour notre programme

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	// Package externe pour la coloration
	"github.com/fatih/color"
)

// Liste des variables globales liées aux flags
var (
	dictPath string // chemin vers le fichier dictionnaire
	target   string // URL / Cible à scanner
	workers  int    // nombre de threads/goroutines en parallèle
	quiet    bool   // mode silencieux (affiche uniquement les 200)
)

// --- AJOUT ---
// On utilise une map pour stocker le nombre de fois où un code HTTP est rencontré.
// Comme on est en contexte concurrent, on protège l’accès avec un Mutex.
var statusCount = make(map[int]int)
var statusMutex sync.Mutex

/*
Dans cette fonction init(), nous déclarons et configurons les différentes options
de ligne de commande (flags) que l'utilisateur pourra renseigner lors de l'exécution
du programme. Nous utilisons pour cela le package standard "flag" de Go.
*/
func init() {
	flag.StringVar(&dictPath, "d", "", "Chemin vers le fichier dictionnaire (ex: git_wordlist.txt)")
	flag.StringVar(&target, "t", "", "Cible à scanner (ex: 127.0.0.1:8000)")
	flag.IntVar(&workers, "w", 1, "Nombre de threads (par défaut : 1)")
	flag.BoolVar(&quiet, "q", false, "Mode silencieux (n’affiche que les 200)")
}

func main() {

	// Lecture des flags
	flag.Parse()

	// Validation des flags (dictionary + target sont obligatoires)
	if dictPath == "" || target == "" {
		fmt.Println("Utilisation :")
		flag.PrintDefaults()
		log.Fatalf("Erreur : il manque --dictionary ou --target.")
	}
	if workers < 1 {
		workers = 1
	}

	// Ouverture du fichier dictionnaire
	file, err := os.Open(dictPath)
	if err != nil {
		log.Fatalf("Erreur : impossible d’ouvrir le fichier dictionnaire : %v", err)
	}
	defer file.Close()

	// Lecture du dictionnaire dans un slice
	var paths []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			paths = append(paths, line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Erreur lors de la lecture du fichier dictionnaire : %v", err)
	}

	// Vérification/formatage de la cible (http:// ou https://)
	finalURL, err := formatTargetURL(target)
	if err != nil {
		log.Fatalf("Erreur sur le format de la cible : %v", err)
	}

	// Affichage des infos de démarrage (sauf en mode quiet)
	if !quiet {
		fmt.Printf("=== Démarrage du scan ===\n")
		fmt.Printf("Cible  : %s\n", finalURL)
		fmt.Printf("Dico   : %s\n", dictPath)
		fmt.Printf("Threads: %d\n", workers)
		fmt.Println("-------------------------")
	}

	// On lance le scan avec goroutines
	start := time.Now()
	results := make(chan string) // Pour collecter et afficher les résultats

	var wg sync.WaitGroup

	// Création des workers
	pathChan := make(chan string)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range pathChan {
				status, err := checkPath(finalURL, p)
				if err != nil {
					// On peut ignorer ou loguer selon le besoin
					continue
				}

				// --- AJOUT ---
				// Mise à jour de la map de comptage des statuts
				statusMutex.Lock()
				statusCount[status]++
				statusMutex.Unlock()

				// En mode quiet, on n’affiche que les 200
				if quiet {
					if status == 200 {
						results <- colorStatus(p, status)
					}
				} else {
					// Sinon, on affiche tous les status, avec couleur
					results <- colorStatus(p, status)
				}
			}
		}()
	}

	// Goroutine pour collecter les résultats et les afficher
	go func() {
		for r := range results {
			fmt.Println(r)
		}
	}()

	// Envoi de chaque path dans le channel pour les workers
	for _, p := range paths {
		pathChan <- p
	}
	close(pathChan)

	// Attente de la fin de toutes les goroutines
	wg.Wait()
	close(results)

	elapsed := time.Since(start)

	if !quiet {
		fmt.Printf("\n--- Scan terminé en %s ---\n", elapsed)

		// --- AJOUT --- Affichage d'un petit résumé des statuts
		fmt.Println("\n--- Résumé des réponses ---")
		statusMutex.Lock()
		for code, count := range statusCount {
			// On peut réutiliser colorStatus() pour colorer, mais ça prend un "path".
			// Du coup, on peut faire un helper rapide :
			line := colorSummary(code, count)
			fmt.Println(line)
		}
		statusMutex.Unlock()
	}
}

// formatTargetURL vérifie si la cible a bien un schéma (http/https).
func formatTargetURL(t string) (string, error) {
	if strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
		return t, nil
	}
	// Sinon, on ajoute "http://" par défaut
	if strings.Contains(t, ":") || strings.Contains(t, ".") {
		return "http://" + t, nil
	}
	return "", errors.New("cible invalide (pas d'URL ni de port détecté)")
}

// checkPath effectue une requête GET sur finalURL + "/" + path
func checkPath(baseURL, path string) (int, error) {
	url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

// colorStatus retourne une chaîne colorée selon le status HTTP (200, 403, 404, autre)
func colorStatus(p string, status int) string {
	switch status {
	case 200:
		return color.New(color.FgGreen).Sprintf("/%s\t%d", p, status)
	case 403:
		return color.New(color.FgHiYellow).Sprintf("/%s\t%d", p, status)
	case 404:
		return color.New(color.FgRed).Sprintf("/%s\t%d", p, status)
	default:
		// Autre couleur
		return color.New(color.FgBlue).Sprintf("/%s\t%d", p, status)
	}
}

// --- AJOUT ---
// colorSummary affiche un code HTTP et un count, colorés.
func colorSummary(code, count int) string {
	switch code {
	case 200:
		return color.New(color.FgGreen).
			Sprintf("HTTP %d -> %d occurrences", code, count)
	case 403:
		return color.New(color.FgHiYellow).
			Sprintf("HTTP %d -> %d occurrences", code, count)
	case 404:
		return color.New(color.FgRed).
			Sprintf("HTTP %d -> %d occurrences", code, count)
	default:
		return color.New(color.FgBlue).
			Sprintf("HTTP %d -> %d occurrences", code, count)
	}
}
