import qdrant_client
from openai import OpenAI

class VectorMemory:
    def __init__(self, collection_name: str):
        self.client = OpenAI()
        self.qdrant = qdrant_client.QdrantClient(host="localhost")
        self.embedding_model = "text-embedding-3-large"
        self.collection_name = collection_name

    def query_qdrant(self, query, vector_name="article", top_k=5):
        # Creates embedding vector from user query
        embedded_query = (
            self.client.embeddings.create(
                input=query,
                model=self.embedding_model,
            )
            .data[0]
            .embedding
        )

        query_results = self.qdrant.search(
            collection_name=self.collection_name,
            query_vector=(vector_name, embedded_query),
            limit=top_k,
        )

        return query_results
    
    def query_docs(query):
        """Query the knowledge base for relevant articles."""
        print(f"Searching knowledge base with query: {query}")
        query_results = query_qdrant(query, collection_name=collection_name)
        output = []

        for i, article in enumerate(query_results):
            title = article.payload["title"]
            text = article.payload["text"]
            url = article.payload["url"]

            output.append((title, text, url))

        if output:
            title, content, _ = output[0]
            response = f"Title: {title}\nContent: {content}"
            truncated_content = re.sub(
                r"\s+", " ", content[:50] + "..." if len(content) > 50 else content
            )
            print("Most relevant article title:", truncated_content)
            return {"response": response}
        else:
            print("No results")
            return {"response": "No results found."}