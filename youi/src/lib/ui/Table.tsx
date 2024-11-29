import { jsx } from "../template";
import { eventBus } from "../event";
import { DynamicIsland } from "./DynamicIsland";

interface TableHeaderProps {
    columns: { label: string; key: string }[];
}

// Define the type for currentSort
type SortState = { field: string; direction: 1 | -1 } | null;
let currentSort: SortState = null;

export const TableHeader = ({ columns }: TableHeaderProps) => {
    // Subscribe to sort events to update currentSort
    eventBus.subscribe("table:sort", (event: { target: HTMLElement }) => {
        const field = event.target.dataset.field;
        // Add null checks and type assertions
        if (!field) return;

        const currentValue = parseInt(event.target.dataset.value ?? "1");

        currentSort = {
            field: field, // Now field is guaranteed to be string
            direction: currentValue === 1 ? -1 : (1 as 1 | -1)
        };

        // currentSort is now guaranteed to be non-null
        event.target.dataset.value = currentSort.direction.toString();
    });

    return (
        <thead>
            <tr>
                {columns.map((column: { label: string; key: string }) => {
                    const sortValue =
                        currentSort && currentSort.field === column.label
                            ? currentSort.direction
                            : 1;

                    return (
                        <th
                            data-trigger="click"
                            data-event="table:sort"
                            data-field={column.label}
                            data-value={sortValue}
                        >
                            {column.label}
                            <i className="material-icons sort-btn">sort</i>
                        </th>
                    );
                })}
                <th>Actions</th>
            </tr>
        </thead>
    );
};

interface TableBodyProps {
    items: any[];
    columns: { key: string; label: string }[];
}

export const TableBody = ({ items, columns }: TableBodyProps) => {
    const tbody = document.createElement("tbody");

    // Helper function to create a table row
    const createTableRow = (item: any) => {
        const row = document.createElement("tr");
        const actionsCell = document.createElement("td");

        // Create and append cells for data
        columns.forEach(({ key }) => {
            const td = document.createElement("td");
            td.textContent = item[key];
            row.appendChild(td);
        });

        // Fix: Append actionsCell immediately, then populate it when Promise resolves
        row.appendChild(actionsCell);
        jsx(DynamicIsland, {
            variant: "contextmenu",
            data: [
                {
                    category: "more_horiz",
                    collections: [
                        {
                            icon: "visibility",
                            label: "View",
                            path: `${window.location.href}/document/${item._id}`
                        },
                        {
                            icon: "delete",
                            label: "Delete",
                            path: `${window.location.href}/delete/${item._id}`
                        }
                    ]
                }
            ]
        }).then((button) => {
            actionsCell.appendChild(button);
        });

        return row;
    };

    // Initial render
    items.forEach((item) => {
        tbody.appendChild(createTableRow(item));
    });

    // Listen for table data updates
    eventBus.subscribe("table:update", (data: { items: any[] }) => {
        tbody.innerHTML = "";
        data.items.forEach((item) => {
            tbody.appendChild(createTableRow(item));
        });
    });

    return tbody;
};

interface TableFooterProps {
    total: number;
    currentPage: number;
}

export const TableFooter = ({ total, currentPage }: TableFooterProps) => {
    const pageSize = 10;
    const totalPages = Math.ceil(total / pageSize);

    eventBus.subscribe("table:paginate", (event: any) => {
        const currentPage = parseInt(event.page || "1", 10);
        if (isNaN(currentPage) || currentPage < 1) return;

        // Update the DOM elements' data-page attributes
        const prevButton =
            event.target.parentNode.querySelector("span:first-child");
        const nextButton =
            event.target.parentNode.querySelector("span:last-child");
        const pageDisplay =
            event.target.parentNode.querySelector("span:nth-child(2)");

        if (prevButton) prevButton.dataset.page = (currentPage - 1).toString();
        if (nextButton) nextButton.dataset.page = (currentPage + 1).toString();
        if (pageDisplay)
            pageDisplay.textContent = `Page ${currentPage} of ${totalPages}`;
    });

    return (
        <tfoot>
            <tr>
                <td>
                    <span
                        data-trigger="click"
                        data-event="table:paginate"
                        data-page={currentPage - 1}
                    >
                        Previous
                    </span>
                    <span>
                        Page {currentPage} of {totalPages}
                    </span>
                    <span
                        data-trigger="click"
                        data-event="table:paginate"
                        data-page={currentPage + 1}
                    >
                        Next
                    </span>
                </td>
            </tr>
        </tfoot>
    );
};

interface TableProps {
    head: JSX.Element;
    body: JSX.Element;
    footer: JSX.Element;
}

export const Table = ({ head, body, footer }: TableProps) => {
    return (
        <table>
            {head}
            {body}
            {footer}
        </table>
    );
};
