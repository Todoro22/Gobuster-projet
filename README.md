# Gobuster-like en Go

**Auteur :** Todoro EQ : YB 
**Date :** 2025-01-24  

---

## 1. Description
Ce projet implémente un outil similaire à **Gobuster** en **Go**.  
Il a pour but de découvrir des répertoires ou des fichiers potentiellement cachés sur un serveur web, en testant des chemins provenant d’un **fichier dictionnaire** (ex. `git_wordlist.txt`).  

---

## 2. Installation & Prérequis
- **Go** installé (Go 1.16 ou version plus récente recommandée).  
- Un éditeur de texte / IDE (Visual Studio Code, GoLand, Vim, etc.).  
- _(Optionnel)_ **Git** pour cloner le dépôt et versionner.

---

## 3. Options disponibles
- **`-dictionary`**  
  Chemin vers le fichier dictionnaire (ex. `git_wordlist.txt`).
  
- **`-target`**  
  URL / hôte à scanner (ex. `http://127.0.0.1:8000`).  
  Peut inclure un port spécifique (ex. `:8080`).

- **`-threads`**  
  Nombre de goroutines (threads) en parallèle (valeur par défaut : **1**).

- **`-quiet`**  
  Mode silencieux, qui **n’affiche que** les résultats ayant un statut HTTP **200**.

---

## 4. Fonctionnement
1. Le programme **lit** chaque ligne du dictionnaire (`git_wordlist.txt`).  
2. Pour chaque mot, il **construit** une URL (ex. `http://localhost:8080/admin`).  
3. Il effectue une **requête HTTP GET** et récupère le **code de statut** (200, 404, 403, etc.).  
4. Un **système de couleurs** met en évidence les différents statuts (vert pour 200, rouge pour 404, etc.).  
5. Un **résumé** final peut être affiché, indiquant le nombre d’occurrences par code HTTP.  

---

## 5. Commande à exécuter
Voici un exemple de commande de lancement :
```bash
go run mainV1.5.go -d git_wordlist.txt -t URL...... -w 5
```

---

## 6. Licence
Ce projet est sous licence Todoro

---

## 7. Avertissement
Ce programme est à titre éducatif et de démonstration uniquement. L'utilisation de cet outil sur des serveurs sans autorisation explicite est illégale. L'auteur décline toute responsabilité pour toute utilisation abusive de ce programme.



