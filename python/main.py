from swarm import Swarm, Agent
from swarm.repl import run_demo_loop

client = Swarm()

def store_memory() -> str:
    

def build_agent(name: str, instructions: str) -> Agent:
    """Builds a component.

    Args:
        name: Name of the agent.
        instructions: Instructions for the agent.
    """
    return Agent(
        name=name,
        instructions=instructions,
        functions=[store_memory, retrieve_memory]
    )

system_agent = Agent(
    name="System Agent",
    functions=[build_agent],
    instructions="You are a system, capable of building your own agents, which you should always do, delegation is key. {username} is the user you are helping.",
)

run_demo_loop(system_agent, stream=True)