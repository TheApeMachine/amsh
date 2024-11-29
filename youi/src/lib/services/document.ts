import menuConfig from "@/menu.json";
import { fetchDocument } from "@/lib/mongo/client";
import { eventBus } from "@/lib/event";
import { stateManager } from "@/lib/state";

// Add type definitions
interface JoinField {
    join: string;
    on: string;
    select: string;
}

// Update interface to match actual menu.json structure
interface MenuConfig {
    category: string;
    collections?: {
        label: string;
        path: string;
        collection: string;
        fields: Array<string | {
            join: string;
            on: string;
            select: string;
        }>;
        requiredFields: string[];
    }[];
}

const createDocumentService = () => {
    const handleStateChange = async (collectionName: string, state: any, params: { id: string }) => {
        try {
            const collection = (menuConfig as MenuConfig[])
                .flatMap((category) => category.collections || [])
                .find((item) => item?.collection === collectionName);

            const joins = collection?.fields?.filter(
                (field): field is JoinField =>
                    typeof field === "object" &&
                    field !== null &&
                    "join" in field &&
                    "on" in field &&
                    "select" in field
            );

            // Build pipeline for single document fetch with joins
            const pipeline: any[] = [
                { $match: { _id: params.id } }
            ];

            // Add joins if they exist
            if (joins?.length) {
                joins.forEach((join: JoinField) => {
                    const [localField, foreignField] = join.on.split(",").map((f: string) => f.trim());
                    pipeline.push({
                        $lookup: {
                            from: join.join,
                            localField: foreignField,
                            foreignField: localField,
                            as: join.join
                        }
                    });
                    pipeline.push({
                        $unwind: {
                            path: `$${join.join}`,
                            preserveNullAndEmptyArrays: true
                        }
                    });
                });

                // Add null check before accessing collection.fields
                if (!collection) {
                    throw new Error(`Collection ${collectionName} not found`);
                }

                // Add a projection stage to specify which fields to include
                pipeline.push({
                    $project: {
                        _id: 1,
                        ...Object.fromEntries(collection.fields.map(field =>
                            typeof field === 'string' ? [field, 1] : [field.join, 1]
                        ))
                    }
                });
            }

            // Update fetchDocument call to use the pipeline
            const document = await fetchDocument(collectionName, params.id, pipeline);

            // Update state
            stateManager.setState({
                [collectionName]: {
                    ...state,
                    items: document,
                    totalPages: 1
                }
            });

            // Trigger re-render
            eventBus.publish("render", collectionName);
        } catch (error) {
            console.error("Error fetching document:", error);
        }
    };

    return {
        handleStateChange,
        getDocument: fetchDocument
    };
};

const documentService = createDocumentService();
export const { handleStateChange, getDocument } = documentService; 