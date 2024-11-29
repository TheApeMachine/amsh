import * as Realm from "realm-web";
import { FromBinary, ToBinary } from "./utils";

// Initialize the Realm app once
const app = new Realm.App({ id: import.meta.env.VITE_REALM_APP_ID });
let user: Realm.User | null = null;

// Login function
const ensureUser = async (): Promise<Realm.User> => {
    if (!user) {
        const credentials = Realm.Credentials.apiKey(import.meta.env.VITE_MONGO_API_KEY);
        user = await app.logIn(credentials);
    }
    return user;
};

const convertBinary = (doc: any): any => {
    if (!doc || typeof doc !== 'object') return doc;

    // Handle arrays recursively
    if (Array.isArray(doc)) {
        return doc.map((item: any): any => {
            if (item?.sub_type === 3 && item?.buffer instanceof Uint8Array) {
                return FromBinary(item);
            }
            return convertBinary(item);
        });
    }

    const converted = { ...doc };

    Object.entries(doc).forEach(([key, value]: any) => {
        if (value?.sub_type === 3 && value?.buffer instanceof Uint8Array) {
            converted[key] = FromBinary(value);
        } else if (value?.$date?.$numberLong) {
            converted[key] = new Date(parseInt(value.$date.$numberLong)).toISOString();
        } else if (value instanceof Date) {
            converted[key] = value.toISOString();
        } else if (value && typeof value === 'object') {
            converted[key] = convertBinary(value);
        }
    });

    return converted;
};

export const fetchCollection = async (
    name: string,
    options: {
        query?: Record<string, any>;
        projection?: string[];
        join?: { from: string; localField: string; foreignField: string; as: string };
        sort?: Record<string, 1 | -1>;
        limit?: number;
        offset?: number;
        count?: boolean;
    } = {}
) => {
    const { query = {}, projection = [], join, sort, limit = 20, offset = 0, count = false } = options;

    const user = await ensureUser();
    const conn = user.mongoClient("mongodb-atlas").db("FanApp");

    // If count is requested, return the total count
    if (count) {
        const pipeline = [
            ...(Object.keys(query).length > 0 ? [{ $match: query }] : []),
            { $count: "total" }
        ];
        const result = await conn.collection(name).aggregate(pipeline);
        return result[0]?.total || 0;
    }

    // Rest of the existing pipeline logic...
    const pipeline: any[] = [];

    if (Object.keys(query).length > 0) {
        pipeline.push({ $match: query });
    }

    if (projection.length > 0) {
        const projectFields = projection.reduce(
            (acc: any, field: string) => ({ ...acc, [field]: 1 }),
            {}
        );
        pipeline.push({ $project: projectFields });
    }

    if (join) {
        pipeline.push({
            $lookup: {
                from: join.from,
                localField: join.localField,
                foreignField: join.foreignField,
                as: join.as
            }
        });
        pipeline.push({ $unwind: `$${join.as}` });
    }

    if (sort) {
        pipeline.push({ $sort: sort });
    }

    pipeline.push({ $skip: offset });
    pipeline.push({ $limit: limit });

    const results = await conn.collection(name).aggregate(pipeline);
    return results.map((doc: any) => convertBinary(doc));
};

export const fetchDocument = async (collectionName: string, id: string, pipeline?: any[]) => {
    const user = await ensureUser();
    const conn = user.mongoClient("mongodb-atlas").db("FanApp");

    if (pipeline) {
        const results = await conn.collection(collectionName).aggregate(pipeline);
        const result = results[0]; // Get first document from pipeline result
        return result ? convertBinary(result) : null;
    }

    // Existing logic for when no pipeline is provided
    const binaryId = ToBinary(id);
    let result = await conn.collection(collectionName).findOne({ _id: binaryId });

    if (!result) {
        console.log('Document not found with binary id, trying original');
        result = await conn.collection(collectionName).findOne({ _id: id });
    }

    if (!result) return null;

    return convertBinary(result);
}; 
