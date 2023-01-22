
from agent import ICMAgent
from frame_stack_env import FrameStack
from ipc_env import GameIpcEnv
from runner import Runner
from utils import get_args

# constants

from stable_baselines3.common.vec_env.vec_transpose import VecTransposeImage
from stable_baselines3.common.vec_env.vec_frame_stack import VecFrameStack
from stable_baselines3.common.vec_env.dummy_vec_env import DummyVecEnv
from stable_baselines3.common.vec_env import VecNormalize
from stable_baselines3.common.env_util import make_vec_env, make_atari_env

if __name__ == '__main__':

    """Argument parsing"""
    args = get_args()

    env = GameIpcEnv()

    """Agent"""
    agent = ICMAgent(args.n_stack, args.num_envs, env.action_space.n, lr=args.lr)

    runner = Runner(agent, env, args.num_envs, args.n_stack, args.rollout_size, args.num_updates,
                    args.max_grad_norm, args.value_coeff, args.entropy_coeff,
                    args.tensorboard, args.log_dir, args.cuda, args.seed)
    runner.train()
