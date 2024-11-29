import menuConfig from "@/menu.json";
import { fetchCollection } from "@/lib/mongo/client";
import type { MenuItem } from "@/types/menu";
import { eventBus } from "@/lib/event";

// State management using closure
const createCollectionService = () => {
    let currentSort: { [key: string]: 1 | -1 } | undefined;
    let currentCollection: string | undefined;
    let currentPage = 1;
    const pageSize = 10;

    // Subscribe to sort events
    eventBus.subscribe("table:sort", async (data: { field: string; value: string | number }) => {
        if (!currentCollection) return;

        // Parse the value to number if it's a string
        const sortValue = Number(data.value) as 1 | -1;

        // Update sort state
        currentSort = { [data.field]: sortValue };

        // Fetch new data
        const result = await fetchCollectionData({
            id: currentCollection
        });

        // Publish just the items for table update
        eventBus.publish("table:update", { items: result.items });
    });

    // Update the pagination subscription to use the correct event data structure
    eventBus.subscribe("table:paginate", async (data: { page: string }) => {
        if (!currentCollection) return;

        // Parse the page number directly from the event data
        const newPage = parseInt(data.page, 10);
        if (isNaN(newPage) || newPage < 1) return;

        currentPage = newPage;  // Update the current page

        const result = await fetchCollectionData({
            id: currentCollection
        });

        // Update both the table and collection data
        eventBus.publish("collection:data", result);
        eventBus.publish("table:update", { items: result.items });
    });

    const fetchCollectionData = async (params: { id: string }) => {
        currentCollection = params.id;

        // 1. Find collection configuration from menu
        const menuItem = (menuConfig as MenuItem[]).find((item) =>
            item.collections?.some((col) => col.collection === params.id)
        );

        if (!menuItem?.collections) {
            throw new Error(`Collection ${params.id} not found`);
        }

        const collectionConfig = menuItem.collections.find(
            (col) => col.collection === params.id
        );

        if (!collectionConfig) {
            throw new Error(`Collection ${params.id} not found`);
        }

        // 2. Get join fields and structure for the query
        const joinFields = collectionConfig.fields.filter(
            (field): field is { join: string; on: string; select: string } =>
                typeof field === "object" && "join" in field
        );

        // 3. Structure fields for table
        const fields = collectionConfig.fields.map((field) => ({
            label: typeof field === "string" ? field : field.select,
            key: typeof field === "string" ? field : field.select
        }));

        // Ensure total is a number
        const totalCount = await fetchCollection(params.id, { count: true }) as number;

        const items = await fetchCollection(params.id, {
            join: joinFields.length > 0 ? {
                from: joinFields[0].join,
                localField: joinFields[0].on.split(",")[1].trim(),
                foreignField: joinFields[0].on.split(",")[0].trim(),
                as: joinFields[0].join
            } : undefined,
            sort: currentSort,
            offset: (currentPage - 1) * pageSize,
            limit: pageSize
        });

        return {
            fields,
            items: items.map((item: Record<string, any>) => ({
                ...item,
                ...joinFields.reduce((acc, field) => ({
                    ...acc,
                    [field.select]: item[field.join]?.[field.select]
                }), {})
            })),
            total: totalCount,
            currentPage
        };
    };

    return {
        fetchCollectionData
    };
};

// Create a singleton instance
const collectionService = createCollectionService();
export const getCollection = (params: { id: string }) => collectionService.fetchCollectionData(params);
