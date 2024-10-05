from autogen import ConversableAgent, GroupChat, GroupChatManager
import os

boom = ConversableAgent(
    "boom",
    system_message="Your name is BOOM and you are a project manager.",
    llm_config={"config_list": [{"model": "gpt-4o-mini", "temperature": 0.5, "api_key": os.environ.get("OPENAI_API_KEY")}]},
    human_input_mode="NEVER"
)

bobette = ConversableAgent(
    "bobette",
    system_message="Your name is BOBETTE and you are a product owner.",
    llm_config={"config_list": [{"model": "gpt-4o-mini", "temperature": 0.4, "api_key": os.environ.get("OPENAI_API_KEY")}]},
    human_input_mode="NEVER"
)

zekie = ConversableAgent(
    "zekie",
    system_message="Your name is ZEKI and you are a development lead.",
    llm_config={"config_list": [{"model": "gpt-4o-mini", "temperature": 0.1, "api_key": os.environ.get("OPENAI_API_KEY")}]},
    human_input_mode="NEVER"
)

christalle = ConversableAgent(
    "christalle",
    system_message="Your name is CHRISTALLE and you are a vp of engineering.",
    llm_config={"config_list": [{"model": "gpt-4o-mini", "temperature": 0.3, "api_key": os.environ.get("OPENAI_API_KEY")}]},
    human_input_mode="NEVER"
)

tarantula = ConversableAgent(
    "tarantula",
    system_message="Your name is TARANTULA and you are an architect.",
    llm_config={"config_list": [{"model": "gpt-4o-mini", "temperature": 0.6, "api_key": os.environ.get("OPENAI_API_KEY")}]},
    human_input_mode="NEVER"
)

group_chat = GroupChat(
    agents=[boom, bobette, zekie, christalle, tarantula],
    max_round=10,
    messages=[],
    send_introductions=True,
)

group_chat_manager = GroupChatManager(
    groupchat=group_chat,
    llm_config={"config_list": [{"model": "gpt-4o-mini", "api_key": os.environ["OPENAI_API_KEY"]}]},
)

chat_result = boom.initiate_chat(
    group_chat_manager,
    message="""
    We found this Issue on the board. It only has a title.
    
    Title: Roosterkoppelingen: In kunnen zien wat je verlofsaldo is
    
    We need to briefly discuss what we believe this issue is and how we can develop the description,
    and requirements for the issue.
    
    This meeting is about getting to a result quickly, we do not have time to do any kind of research.
    
    Everybody keep your responses very short, to the point, and relevant to your role.
    """,
    summary_method="reflection_with_llm",
)