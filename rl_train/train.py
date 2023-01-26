import gym.vector

from ipc_env import GameIpcEnv
from stable_baselines3 import PPO
from stable_baselines3 import A2C
from stable_baselines3 import DQN
from stable_baselines3.common.callbacks import CallbackList, CheckpointCallback, EvalCallback
from stable_baselines3.common.type_aliases import TensorDict
from stable_baselines3.common.torch_layers import BaseFeaturesExtractor, NatureCNN
from frame_stack_env import FrameStack
from stable_baselines3.common.preprocessing import get_flattened_obs_dim, is_image_space
import torch.nn as nn
import torch as th
from gym import spaces

from sb3_contrib import RecurrentPPO


class ImageFeatureExtractor(BaseFeaturesExtractor):

    def __init__(
        self,
        observation_space: spaces.Box,
        features_dim: int = 512,
    ) -> None:
        super().__init__(observation_space, features_dim)

        n_input_channels = observation_space.shape[0]
        self.cnn = nn.Sequential(
            nn.Conv2d(n_input_channels, 32, kernel_size=5, stride=2, padding=0),
            nn.ReLU(),
            nn.MaxPool2d(kernel_size=3, stride=2),

            nn.Conv2d(32, 64, kernel_size=3, stride=2, padding=0),
            nn.ReLU(),
            nn.MaxPool2d(kernel_size=3, stride=2),

            nn.Conv2d(64, 128, kernel_size=3, stride=1, padding=0),
            nn.ReLU(),
            nn.MaxPool2d(kernel_size=3, stride=2),

            nn.Flatten(),
        )

        # Compute shape by doing one forward pass
        with th.no_grad():
            n_flatten = self.cnn(th.as_tensor(observation_space.sample()[None]).float()).shape[1]

        self.linear = nn.Sequential(nn.Linear(n_flatten, features_dim), nn.ReLU())

    def forward(self, observations: th.Tensor) -> th.Tensor:
        return self.linear(self.cnn(observations))


class CombinedExtractor(BaseFeaturesExtractor):
    def __init__(self, observation_space: gym.spaces.Dict, cnn_output_dim: int = 32):
        super().__init__(observation_space, features_dim=1)

        self.prev_state = None
        extractors = {}

        total_concat_size = 0
        for key, subspace in observation_space.spaces.items():
            if is_image_space(subspace):
                n_input_channels = subspace.shape[0]
                network = ImageFeatureExtractor(subspace, features_dim=cnn_output_dim)

                extractors[key] = network
                total_concat_size += cnn_output_dim
            else:
                # The observation key is a vector, flatten it if needed
                text_output_dims = 32
                extractors[key] = nn.Sequential(
                    nn.Flatten(),
                    nn.Linear(get_flattened_obs_dim(subspace), 256),
                    nn.ReLU(),
                    nn.Linear(256, text_output_dims),
                )
                total_concat_size += text_output_dims

        self.extractors = nn.ModuleDict(extractors)

        # Update the features dim manually
        self._features_dim = total_concat_size

        self.bn = nn.BatchNorm1d(total_concat_size)


    def forward(self, observations: TensorDict) -> th.Tensor:
        encoded_tensor_list = []

        for key, extractor in self.extractors.items():
            encoded_tensor_list.append(extractor(observations[key]))
        catobs = th.cat(encoded_tensor_list, dim=1)

        catobs = self.bn(catobs)


        return catobs


def train():

    env = GameIpcEnv()

    #env = FrameStack(env, 3)

    checkpoint_callback = CheckpointCallback(
        save_freq=100000,
        save_path="./logs/",
        name_prefix="rl_model",
        save_replay_buffer=False,
        save_vecnormalize=False,
    )
    policy_kwargs = dict(
        features_extractor_class=CombinedExtractor,
        features_extractor_kwargs=dict(),
    )
    model = RecurrentPPO("MultiInputLstmPolicy", learning_rate=0.0001, policy_kwargs=policy_kwargs, env=env, verbose=1, tensorboard_log="./logs/")

    model.learn(total_timesteps=50000000, callback=checkpoint_callback)

    vec_env = model.get_env()
    obs = vec_env.reset()
    lstm_states = None
    for i in range(10000):
        action, lstm_states = model.predict(obs, state=lstm_states, deterministic=True)
        obs, reward, done, info = vec_env.step(action)

    env.close()


# Press the green button in the gutter to run the script.
if __name__ == '__main__':
    train()
