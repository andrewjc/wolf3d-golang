from ipc_env import GameIpcEnv


def train():

    from stable_baselines3 import PPO

    env = GameIpcEnv()

    model = PPO("MlpPolicy", env, verbose=1)
    model.learn(total_timesteps=500000)

    vec_env = model.get_env()
    obs = vec_env.reset()
    for i in range(1000):
        action, _states = model.predict(obs, deterministic=True)
        obs, reward, done, info = vec_env.step(action)
        vec_env.render()
        # VecEnv resets automatically
        # if done:
        #   obs = env.reset()

    env.close()


# Press the green button in the gutter to run the script.
if __name__ == '__main__':
    train()
