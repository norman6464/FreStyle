import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight';
import { ReactNodeViewRenderer, type ReactNodeViewProps } from '@tiptap/react';
import type { ComponentType } from 'react';
import CodeBlockNodeView from '../components/CodeBlockNodeView';

export const CodeBlock = CodeBlockLowlight.extend({
  addNodeView() {
    // CodeBlockNodeView は ReactNodeViewProps<HTMLElement> 拡張型なので、
    // ReactNodeViewRenderer の厳密シグネチャに合わせて cast する。
    return ReactNodeViewRenderer(
      CodeBlockNodeView as ComponentType<ReactNodeViewProps<HTMLElement>>,
    );
  },
});
