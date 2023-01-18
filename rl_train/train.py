import gym.vector

from ipc_env import GameIpcEnv
from stable_baselines3 import PPO
from stable_baselines3 import A2C
from stable_baselines3.common.callbacks import CallbackList, CheckpointCallback, EvalCallback

from frame_stack_env import FrameStack

def train():

    env = GameIpcEnv()

    env = FrameStack(env, 4)

    checkpoint_callback = CheckpointCallback(
        save_freq=100000,
        save_path="./logs/",
        name_prefix="rl_model",
        save_replay_buffer=False,
        save_vecnormalize=False,
    )
    callback = CallbackList([checkpoint_callback])
    model = A2C('MlpPolicy', env, verbose=1, tensorboard_log="./logs/")

    model.learn(total_timesteps=50000000, callback=checkpoint_callback)

    vec_env = model.get_env()
    obs = vec_env.reset()
    for i in range(10000):
        action, _states = model.predict(obs, deterministic=True)
        obs, reward, done, info = vec_env.step(action)

    env.close()


# Press the green button in the gutter to run the script.
if __name__ == '__main__':
    train()
