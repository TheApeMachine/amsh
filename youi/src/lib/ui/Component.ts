import { onMount } from "../lifecycle";
import { loader } from "../loader";

/** Ref type for referencing HTML elements */
export type Ref = { current: HTMLElement | null };

/** Creates a new Ref */
export const createRef = (): Ref => ({ current: null });

/** Configuration interface for Component */
interface ComponentConfig<Props = any> {
    loader?: Record<string, {
        url: string;
        method: string;
    }>;
    loading?: () => Promise<Node | JSX.Element>;
    error?: (error: any) => Promise<Node | JSX.Element>;
    effect?: (data: any) => void;
    render: (props: Props) => Promise<Node | JSX.Element>;
}

/** Creator interface for Component */
type ComponentCreator = {
    create: <Props>(
        render: (props: Props) => Promise<Node | JSX.Element> | Node | JSX.Element
    ) => (props: Props) => JSX.Element;
};

/** Component utility */
export const Component = Object.assign(
    <Props = any>(config: ComponentConfig<Props>) => {
        return (Component as ComponentCreator).create<Props>(
            async (props: Props): Promise<Node | JSX.Element> => {
                try {
                    if (config.loader) {
                        if (config.loading) {
                            return await config.loading();
                        }

                        const { state, results } = await loader(config.loader);
                        
                        if (state === "error") {
                            throw results;
                        }

                        const propsWithData = {
                            ...props,
                            data: results
                        };

                        const renderedElement = await config.render(propsWithData);

                        if (config.effect && renderedElement instanceof HTMLElement) {
                            onMount(renderedElement, () => {
                                config.effect!(propsWithData.data);
                            });
                        }

                        return renderedElement;
                    }

                    const result = config.render(props);
                    return result instanceof Promise ? await result : result;
                } catch (error: unknown) {
                    const errorMessage = error instanceof Error ?
                        error.message :
                        'An unknown error occurred';

                    if (config.error) {
                        const errorElement = await config.error(error);
                        return errorElement;
                    } else {
                        return document.createTextNode(`Error: ${errorMessage}`);
                    }
                }
            }
        );
    },
    {
        create: <Props>(
            render: (props: Props) => Promise<Node | JSX.Element> | Node | JSX.Element
        ) => (props: Props): JSX.Element => {
            return render(props) as JSX.Element;
        }
    }
);