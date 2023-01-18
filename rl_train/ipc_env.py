import base64
import json
import time

import gym
import numpy
import numpy as np
import socket
import cv2
import pandas
from gym.utils import seeding
from gym import utils


class GameIpcEnv(gym.Env, utils.EzPickle):
    def __init__(self):
        utils.EzPickle.__init__(self)
        self._seed(seed=time.time_ns())
        self.episodeNumber = 0
        self.is_connected = None
        self.action_space = gym.spaces.Discrete(7)
        self.IMG_WIDTH = 64
        self.IMG_HEIGHT = 64
        self.num_envs = 1

        self.observation_space = gym.spaces.Box(low=0, high=1, shape=(self.IMG_WIDTH, self.IMG_HEIGHT, 1), dtype=np.float32)
        self.connect()

    def reset(self):
        print("reset")
        self.sendMessage(13, b"reset")
        msgType, msgData = self.readMessageReply()
        if msgType == 14:
            return self.get_observation()

        return None, None


    def _seed(self, seed=None):
        self.np_random, seed1 = seeding.np_random(seed)
        return [seed1]

    def get_action_meanings(self):
        ACTION_MEANING = {
            0: "NOOP",
            1: "FORWARD",
            2: "BACKWARD",
            3: "STRAFE_LEFT",
            4: "STRATE_RIGHT",
            5: "TURN_LEFT",
            6: "TURN_RIGHT",
        }
        return [ACTION_MEANING[i] for i in range(0, self.action_space.n)]

    def step(self, action):
        actionResult = self.sendIpcAction(action)

        if actionResult is None:
            return None, 0.0, True, {}

        obs = actionResult['Observation']
        # base64 decode
        obs = base64.b64decode(obs)

        obs = self.imgFromStream(obs)

        reward = actionResult['Reward']
        done = actionResult['Done']

        info = dict()
        info['episode'] = dict()
        if done:
            info['episode']['r'] = reward
            self.episodeNumber += 1

        return obs, reward, done, {}

    def render(self, **kwargs) -> None:
        print("render")

    def connect(self):
        # Connect via unix socket to game process
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.connect("/tmp/wolf3d_ipc_player.sock")

        # if on windows used a named pipe instead
        #self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        #self.sock.connect("\\.\pipe\wolf3d_ipc_player.sock")

        self.is_connected = True

        # begin handshake
        handshake = self.performHandshake()

        if not handshake:
            self.is_connected = False
            self.sock.close()
            print("Handshake failed")
            return

        status = self.sendMessage(16, b"begin control")
        if status:
            msgType, msgData = self.readMessageReply()


    def performHandshake(self):
        status = self.readIpcHandshake()
        if status:
            # got ipc handshake
            status = self.writeIpcHandshakeReply()
            if status:
                status = self.readIpcMsgLenReply()
                self.setMaxMessageLength(status)

                if status:
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

                    #print(f"Incoming IPC Msg Length: msgLen={msgLen}")
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
                        return msgLen
                except socket.timeout:
                    # Socket timeout, continue the loop
                    continue

    def sendMessage(self, msgType, msgData):
        #print(f"Sending message: type={msgType}, data={msgData}")

        try:
            # encode msgType and msgData into json
            b64Msgdata = base64.b64encode(msgData).decode('utf-8')
            msg = json.dumps({"MsgType": msgType, "Data": b64Msgdata})
            bMessage = msg

            # utf8 encode the string
            bMessage = bMessage.encode("utf-8")


            #bMessage = msgType.to_bytes(4, byteorder='big') + msgData

            msgLength = len(bMessage)

            # send msgType as a series of bytes
            try:
                self.sock.send(msgLength.to_bytes(4, byteorder='big'))
                self.sock.send(bMessage)
            except BrokenPipeError:
                print("Broken pipe error")
                return False
            except Exception as e:
                print("Exception: " + str(e))
                return False

            return True
        except socket.timeout:
            # Socket timeout, continue the loop
            return

    def readMessageReplyBytes(self):
        incomingMsgLen = self.readIpcIncomingMsgLenReply()
        if incomingMsgLen:
            try:
                # Receive data using recv_into
                data = bytearray(incomingMsgLen)
                view = memoryview(data)
                while incomingMsgLen:
                    nbytes = self.sock.recv_into(view, incomingMsgLen)
                    view = view[nbytes:]
                    incomingMsgLen -= nbytes


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
        else:
            return None, None

    def readMessageReply(self):
        incomingMsgLen = self.readIpcIncomingMsgLenReply()
        if incomingMsgLen:
            try:
                # Receive data
                data = bytearray(incomingMsgLen)
                view = memoryview(data)
                while incomingMsgLen:
                    nbytes = self.sock.recv_into(view, incomingMsgLen)
                    view = view[nbytes:]
                    incomingMsgLen -= nbytes

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
        else:
            return None, None

    def get_observation(self):
        success = self.sendMessage(18, b"get observation")
        if success:
            msgType, msgData = self.readMessageReplyBytes()
            if msgType == 19:
                msgReplyObj = json.loads(msgData)
                # base64 decode
                obs = base64.b64decode(msgReplyObj['Observation'])
                img = self.imgFromStream(obs)

                return img

        else:
            return None

    def sendIpcAction(self, action):
        bAction = numpy.int64(action).tobytes()

        self.sendMessage(20, bAction)

        msgType, msgReply = self.readMessageReplyBytes()
        if msgReply:
            msgReplyObj = json.loads(msgReply)
            #print(f"Action reply: {msgReplyObj}")
            return msgReplyObj
        else:
            return None


    def setMaxMessageLength(self, maxLen):
        self.maxMessageLen = maxLen

    def imgFromStream(self, msgData):
        # jpeg decompress msgData into numpy array
        img = cv2.imdecode(np.frombuffer(msgData, np.uint8), cv2.IMREAD_GRAYSCALE)

        # resize to 84x84
        img = cv2.resize(img, (self.IMG_WIDTH, self.IMG_HEIGHT))

        # normalize to 0-1
        img = img / 255.0

        img = img.reshape(self.IMG_WIDTH, self.IMG_HEIGHT, 1)
        return img


