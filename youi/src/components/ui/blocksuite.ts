import '@blocksuite/presets/themes/affine.css';
import { AffineSchemas } from '@blocksuite/blocks';
import { AffineEditorContainer } from '@blocksuite/presets';
import { Doc, Schema } from '@blocksuite/store';
import { DocCollection, Text } from '@blocksuite/store';
import { IndexeddbPersistence } from 'y-indexeddb';

class YouiBlocksuite extends HTMLElement {
    shadowRoot: ShadowRoot;
    template: HTMLTemplateElement;
    schema!: Schema;
    collection!: DocCollection;
    doc!: Doc;
    editor!: AffineEditorContainer;

    constructor() {
        super();
        this.template = document.createElement("template");
        this.template.innerHTML = `
        <style>
            :host {
            --youi-unit: 1rem;
            --youi-unit-2: 2rem;
            --youi-radius: 0.25rem;
            --youi-border-width: 2px;
            --youi-muted: #ccc;
            display: flex;
            align-items: center;
            justify-content: center;
            border: var(--youi-border-width, 2px) solid var(--youi-muted, #ccc);
            border-radius: var(--youi-radius, 0.25rem);
            width: 100%;
            height: 100%;
            }

            affine-editor-container {
            width: 100%;
            height: 100%;
            }
        </style>
        `;
        this.shadowRoot = this.attachShadow({ mode: "open" });
    }

    connectedCallback() {
        this.shadowRoot.appendChild(this.template.content.cloneNode(true));

        // Initialize schema, collection, and document
        this.schema = new Schema().register(AffineSchemas);
        this.collection = new DocCollection({ schema: this.schema });
        this.collection.meta.initialize();
        this.doc = this.collection.createDoc();

        // Create the document and load initial content
        this.createDoc();

        // Initialize editor and append to shadow DOM
        this.editor = new AffineEditorContainer();
        this.editor.doc = this.doc;
        this.shadowRoot.append(this.editor);
    }

    createDoc() {
        // Setting up persistence
        new IndexeddbPersistence('provider-demo', this.doc.spaceDoc);

        // Load and create initial blocks
        this.doc.load(() => {
            const pageBlockId = this.doc.addBlock('affine:page', {
                title: new Text('Test'),
            });
            this.doc.addBlock('affine:surface', {}, pageBlockId);
            const noteId = this.doc.addBlock('affine:note', {}, pageBlockId);
            this.doc.addBlock(
                'affine:paragraph',
                { text: new Text('Hello World!') },
                noteId
            );
        });
    }
}

customElements.define("youi-blocksuite", YouiBlocksuite);
