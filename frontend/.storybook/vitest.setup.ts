import { beforeAll } from 'vitest';
import { setProjectAnnotations } from '@storybook/react-vite';
import * as a11yAddonAnnotations from '@storybook/addon-a11y/preview';
import * as previewAnnotations from './preview';

// story をテストとして走らせるときにも preview の設定（グローバル CSS・書体）と
// a11y 検査を効かせる。登録しないと addon を有効にしていても検査は 1 度も走らず、
// parameters.a11y の設定が黙って無視される。
const project = setProjectAnnotations([a11yAddonAnnotations, previewAnnotations]);

beforeAll(project.beforeAll);
