export interface Prompt {
    session_id: string;
    system: string[];
    assistant: string[];
    tool: string[];
    function: string[];
    user: string[];
}

export interface Agent {
    id: string;
    type: string;
    prompt?: Prompt;
}

export interface Team {
    id: string;
    name: string;
    lead?: Agent;
    agents: Agent[];
    prompt?: Prompt;
    response: string;
}

export interface Chunk {
    session_id: string;
    sequence_id: string;
    iteration: number;
    team?: Team;
    agent?: Agent;
    response: string;
}
