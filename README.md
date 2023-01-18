# Wolfenstein 3D in GoLang

Welcome to the Navigating Raycasting Worlds project! This project aims to provide a custom raycasting game engine written in GoLang, reminiscent of the classic Wolf3D game, and to explore the use of reinforcement learning techniques to train artificial agents to navigate through randomly generated maps.

![Screenshot](assets/screenshot.png?raw=true "Game Screenshot")

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
