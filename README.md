MasonJar is UI for the Nertz card game with a Go backend.

From the beginning in a 3 player scenario
1) A logs in to MasonJar then server sends a message with all the players
    - Ready => generates the board and signals to the server that the user is ready
    - Watch => doesn't block the game from starting and tries to watch the lake
