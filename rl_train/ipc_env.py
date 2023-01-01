import gym
import numpy as np
import socket

class GameIpcEnv(gym.Env):
    def __init__(self):
        self.is_connected = None
        self.action_space = gym.spaces.Discrete(4)
        self.observation_space = gym.spaces.Box(low=0, high=255, shape=(84, 84, 1), dtype=np.uint8)
        self.connect()

    def step(self, action):
        pass

    def connect(self):
        # Connect via unix socket to game process
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.connect("/tmp/wolf3d_ipc_player.sock")

        self.is_connected = True

        # begin handshake
        status = self.readIpcHandshake()
        if status:
            # got ipc handshake
            self.writeIpcHandshakeReply()

        while True:
            try:
                # Receive data
                data = self.sock.recv(1024)
                if not data:
                    # If data is not received, the connection is probably closed
                    break
                print(data)
            except socket.timeout:
                # Socket timeout, continue the loop
                continue

    def readIpcHandshake(self):
        try:
            # Receive data
            data = self.sock.recv(1024)
            if not data:
                # If data is not received, the connection is probably closed
                return False
            else:
                proto_version = data[0]
                proto_encrypted = data[1]
                print(f"IPC Handshake: version={proto_version}, encrypted={proto_encrypted}")
                return True
        except socket.timeout:
            # Socket timeout, continue the loop
            return False

    def writeIpcHandshakeReply(self):
        try:
            # Receive data
            self.sock.send(b"0x0")
        except socket.timeout:
            # Socket timeout, continue the loop
            return


