import gym
import numpy as np
import socket
import cv2



class GameIpcEnv(gym.Env):
    def __init__(self):
        self.is_connected = None
        self.action_space = gym.spaces.Discrete(4)
        self.IMG_WIDTH = 64
        self.IMG_HEIGHT = 64

        self.observation_space = gym.spaces.Box(low=0, high=255, shape=(self.IMG_WIDTH, self.IMG_HEIGHT, 3), dtype=np.uint8)
        self.connect()

    def reset(self):
        obs = self.get_observation()

        return obs

    def step(self, action):

        self.sendIpcAction(action)

        obs = self.get_observation()
        reward = 0
        done = False
        info = {
            "action": action
        }

        return obs, reward, done, info

    def connect(self):
        # Connect via unix socket to game process
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.connect("/tmp/wolf3d_ipc_player.sock")

        self.is_connected = True

        # begin handshake
        handshake = self.performHandshake()

        if not handshake:
            self.is_connected = False
            self.sock.close()
            print("Handshake failed")
            return

        print("Handshake successful, requesting control.")
        status = self.sendMessage(100, b"begin control")
        if status:
            msgType, msgData = self.readMessageReply()
            if msgType:
                print(f"Control granted.")


    def performHandshake(self):
        status = self.readIpcHandshake()
        if status:
            # got ipc handshake
            status = self.writeIpcHandshakeReply()
            if status:
                status = self.readIpcMsgLenReply()

                if status:
                    # msg len reply
                    status = self.writeIpcMsgLenReply()

                    return status

        return False

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
            # send decimal 0 as bytes
            self.sock.sendall(b'\x00')
            return True
        except socket.timeout:
            # Socket timeout, continue the loop
            return

    def writeIpcMsgLenReply(self):
        try:
            # send decimal 0 as bytes
            self.sock.sendall(b'\x00')
            return True
        except socket.timeout:
            # Socket timeout, continue the loop
            return

    def readIpcIncomingMsgLenReply(self):
        while True:
            try:
                # Receive data
                data = self.sock.recv(4)
                if not data:
                    # If data is not received, the connection is probably closed
                    return None
                else:
                    # convert binary to decimal
                    msgLen = int.from_bytes(data, byteorder='big')
                    print(f"Incoming IPC Msg Length: msgLen={msgLen}")
                    return msgLen
            except socket.timeout:
                # Socket timeout, continue the loop
                continue

    def readIpcMsgLenReply(self):
        incomingMsgLen = self.readIpcIncomingMsgLenReply()
        if incomingMsgLen:
            while True:
                try:
                    # Receive data
                    data = self.sock.recv(incomingMsgLen)
                    if not data:
                        # If data is not received, the connection is probably closed
                        return None
                    else:
                        # convert binary to decimal
                        msgLen = int.from_bytes(data, byteorder='big')
                        print(f"Set Max Msg Length: msgLen={msgLen}")
                        return msgLen
                except socket.timeout:
                    # Socket timeout, continue the loop
                    continue

    def sendMessage(self, msgType, msgData):
        print(f"Sending message: type={msgType}, data={msgData}")

        try:
            bMessage = msgType.to_bytes(4, byteorder='big') + msgData
            msgLength = len(bMessage)


            # send msgType as a series of bytes
            self.sock.send(msgLength.to_bytes(4, byteorder='big'))

            self.sock.send(bMessage)

            return True
        except socket.timeout:
            # Socket timeout, continue the loop
            return

    def readMessageReplyBytes(self):
        incomingMsgLen = self.readIpcIncomingMsgLenReply()
        if incomingMsgLen:
            try:
                # Receive data
                data = self.sock.recv(incomingMsgLen)
                if not data:
                    # If data is not received, the connection is probably closed
                    return None, None
                else:
                    # read first 4 bytes into msgType
                    msgType = int.from_bytes(data[0:4], byteorder='big')
                    msgData = data[4:]
                    return msgType, msgData
            except Exception as e:
                # Socket timeout, continue the loop
                print("readMessageReply failed:", e)
                return None, None

    def readMessageReply(self):
        incomingMsgLen = self.readIpcIncomingMsgLenReply()
        if incomingMsgLen:
            try:
                # Receive data
                data = self.sock.recv(incomingMsgLen)
                if not data:
                    # If data is not received, the connection is probably closed
                    return None, None
                else:
                    # read first 4 bytes into msgType
                    msgType = int.from_bytes(data[0:4], byteorder='big')
                    msgData = data[4:]

                    # read msgData as string
                    msgData = msgData.decode("utf-8")

                    return msgType, msgData
            except Exception as e:
                # Socket timeout, continue the loop
                print("readMessageReply failed:", e)
                return None, None

    def get_observation(self):
        success = self.sendMessage(102, b"get observation")
        if success:
            msgType, msgData = self.readMessageReplyBytes()
            if msgType == 103:
                # jpeg decompress msgData into numpy array
                img = cv2.imdecode(np.frombuffer(msgData, np.uint8), cv2.IMREAD_COLOR)

                # resize to 84x84
                img = cv2.resize(img, (self.IMG_WIDTH, self.IMG_HEIGHT))

                # normalize to 0-1
                img = img / 255.0

                img = img.reshape(self.IMG_WIDTH, self.IMG_HEIGHT, 3)

                return img

        else:
            return None

    def sendIpcAction(self, action):
        bAction = action.tobytes()


        self.sendMessage(200, bAction)

        msgReply = self.readMessageReply()
        if msgReply:
            print(f"Action reply: {msgReply}")


