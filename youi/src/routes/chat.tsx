import './styles.css';
import { jsx } from '@/lib/template';

import { registerDragonSupport } from '@lexical/dragon';
import { createEmptyHistoryState, registerHistory } from '@lexical/history';
import { HeadingNode, QuoteNode, registerRichText } from '@lexical/rich-text';
import { mergeRegister } from '@lexical/utils';
import { createEditor } from 'lexical';

export const render = async () => {

    const editorRef = document.getElementById('lexical-editor');
    const stateRef = document.getElementById(
        'lexical-state',
    ) as HTMLTextAreaElement;

    const initialConfig = {
        namespace: 'Vanilla JS Demo',
        // Register nodes specific for @lexical/rich-text
        nodes: [HeadingNode, QuoteNode],
        onError: (error: Error) => {
            throw error;
        },
        theme: {
            // Adding styling to Quote node, see styles.css
            quote: 'PlaygroundEditorTheme__quote',
        },
    };
    const editor = createEditor(initialConfig);
    editor.setRootElement(editorRef);

    // Registring Plugins
    mergeRegister(
        registerRichText(editor),
        registerDragonSupport(editor),
        registerHistory(editor, createEmptyHistoryState(), 300),
    );

    editor.registerUpdateListener(({ editorState }) => {
        stateRef!.value = JSON.stringify(editorState.toJSON(), undefined, 2);
    });

    return Transition(match(await loader({}), {
        loading: () => {
            return html`<div>Loading...</div>`
        },
        error: (error: any) => {
            return html`<div>Error: ${error}</div>`
        },
        success: (_: any) => (
            <div>
            <h1>Lexical Basic - Vanilla JS</h1>
            <div class="editor-wrapper">
              <div id="lexical-editor" contenteditable></div>
            </div>
            <h4>Editor state:</h4>
            <textarea id="lexical-state"></textarea>
          </div>
    )}), {
        enter: sequence(blurIn),
        exit: sequence(blurOut)
    })

}
