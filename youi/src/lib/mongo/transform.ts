import { fetchCollection } from "./client";

interface TransformData {
    items: any[];
    page?: number;
}

interface JoinConfig {
    join: string;
    on: string;
}

interface SortOptions {
    field: string;
    order: "asc" | "desc";
}

export const transform = {
    withJoins: (joins: JoinConfig[]) =>
        async (data: TransformData) => {
            if (!joins.length) return data;

            const joinConfig = joins.map(field => ({
                from: field.join,
                localField: field.on.split(",")[1],
                foreignField: field.on.split(",")[0],
                as: field.join
            }));

            return {
                ...data,
                items: await Promise.all(data.items.map(async (item: any) => {
                    const joinedData = await Promise.all(
                        joinConfig.map(async (config: { from: string, localField: string, foreignField: string }) => {
                            const joined = await fetchCollection(config.from, {
                                query: { [config.foreignField]: item[config.localField] }
                            });
                            return { [config.from]: joined[0] };
                        })
                    );
                    return { ...item, ...Object.assign({}, ...joinedData) };
                }))
            };
        },

    withPagination: (pageSize = 10) =>
        (data: TransformData) => ({
            ...data,
            pagination: {
                current: data.page ?? 1,
                pageSize,
                total: Math.ceil(data.items.length / pageSize)
            },
            items: data.items.slice(
                ((data.page ?? 1) - 1) * pageSize,
                (data.page ?? 1) * pageSize
            )
        }),

    withSort: (options: SortOptions) =>
        (data: TransformData) => ({
            ...data,
            sort: options,
            items: [...data.items].sort((a, b) => {
                return (a[options.field] - b[options.field]) * (options.order === "asc" ? 1 : -1);
            })
        })
};
