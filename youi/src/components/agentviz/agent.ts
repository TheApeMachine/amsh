export type Agent = { team: number; agent: number };

export type Conversation = {
    id: number;
    from: Agent;
    to: Agent;
    message: string;
    sentiment: 'positive' | 'negative';
    threadId: number;
};
