
from agent import ICMAgent
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

    make_env = lambda: GameIpcEnv()
    env = make_vec_env(make_env,
                       n_envs=1,
                       monitor_dir=None,
                       wrapper_class=None,
                       env_kwargs=None,
                       vec_env_cls=None,
                       vec_env_kwargs=None,
                       monitor_kwargs=None)

    env = VecTransposeImage(env)

    env = VecFrameStack(env, n_stack=args.n_stack)

    """Agent"""
    agent = ICMAgent(args.n_stack, args.num_envs, env.action_space.n, lr=args.lr)



    runner = Runner(agent, env, args.num_envs, args.n_stack, args.rollout_size, args.num_updates,
                    args.max_grad_norm, args.value_coeff, args.entropy_coeff,
                    args.tensorboard, args.log_dir, args.cuda, args.seed)
    runner.train()
