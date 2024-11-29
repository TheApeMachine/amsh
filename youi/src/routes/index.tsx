import { Component } from "@/lib/ui/Component";
import { jsx } from "@/lib/template";
import { Layout } from "@/lib/ui/layout/Layout";

/** Main Render Function */
export const render = Component({
    loader: {
        "users": {
            url: "https://jsonplaceholder.typicode.com/users",
            method: "GET"
        }
    },
    loading: async () => {
        return <div>Loading...</div>;
    },
    error: async (error: any) => {
        return <div>Error: {error}</div>;
    },
    effect: (data: any) => {
        const usersDiv = document.getElementById("users");
        if (usersDiv) {
            usersDiv.innerHTML = data.users[0].name;
        }
    },
    render: async () => {
        return <Layout>{<div id="users"></div>}</Layout>;
    }
});
