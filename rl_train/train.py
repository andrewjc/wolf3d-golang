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

class CombinedExtractor(BaseFeaturesExtractor):
    """
    Combined feature extractor for Dict observation spaces.
    Builds a feature extractor for each key of the space. Input from each space
    is fed through a separate submodule (CNN or MLP, depending on input shape),
    the output features are concatenated and fed through additional MLP network ("combined").

    :param observation_space:
    :param cnn_output_dim: Number of features to output from each CNN submodule(s). Defaults to
        256 to avoid exploding network sizes.
    """

    def __init__(self, observation_space: gym.spaces.Dict, cnn_output_dim: int = 256):
        # TODO we do not know features-dim here before going over all the items, so put something there. This is dirty!
        super().__init__(observation_space, features_dim=1)

        extractors = {}

        total_concat_size = 0
        for key, subspace in observation_space.spaces.items():
            if is_image_space(subspace):
                n_input_channels = subspace.shape[0]
                network = NatureCNN(subspace, features_dim=cnn_output_dim)

                extractors[key] = network
                total_concat_size += cnn_output_dim
            else:
                # The observation key is a vector, flatten it if needed
                extractors[key] = nn.Sequential(
                    nn.Flatten(),
                    nn.Linear(get_flattened_obs_dim(subspace), 256),
                    nn.ReLU(),
                    nn.Linear(256, get_flattened_obs_dim(subspace)),
                )
                total_concat_size += get_flattened_obs_dim(subspace)

        self.extractors = nn.ModuleDict(extractors)

        # Update the features dim manually
        self._features_dim = total_concat_size

    def forward(self, observations: TensorDict) -> th.Tensor:
        encoded_tensor_list = []

        for key, extractor in self.extractors.items():
            encoded_tensor_list.append(extractor(observations[key]))
        return th.cat(encoded_tensor_list, dim=1)


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
    model = DQN("MultiInputPolicy", policy_kwargs=policy_kwargs, env=env, verbose=1, tensorboard_log="./logs/")

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
