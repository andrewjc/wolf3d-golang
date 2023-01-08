# Wolfenstein 3D in GoLang

Welcome to the Wolfenstein 3D in GoLang project! This project aims to provide a reimplementation of the classic first-person shooter game Wolfenstein 3D, written in the Go programming language, and to explore the use of machine learning techniques to train artificial neural networks to play the game.

![Screenshot](assets/game_wnd.png?raw=true "Game Screenshot")

## Getting Started

To get started with the project, you will need to have the following dependencies installed:

- GoLang: You can download the latest version of GoLang from the [official website](https://golang.org/).

- Python: You will need to have Python installed to run the neural network trainer. You can download Python from the [official website](https://www.python.org/).

Once you have these dependencies installed, you can clone the project repository and build the GoLang code by running the following commands:

```
git clone https://github.com/andrewjc/wolf3d-golang.git
cd wolfenstein-3d-golang
make build-game
```

You can then run the game by running the following command:

```
./build/game
```

This will start the game env and open an IPC interface allowing the player and enemy character(s) to be controlled by the neural network trainer. You can then run the neural network trainer by running the following command:

```
make train
```
