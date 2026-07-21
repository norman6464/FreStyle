import { useDispatch, useSelector } from 'react-redux';
import type { TypedUseSelectorHook } from 'react-redux';
import type { RootState, AppDispatch } from '@/app/store';

/*
 * 型付きの Redux hooks。
 *
 * なぜ shared に置くのか:
 * store 本体（configureStore / RootState）は各 slice を集約する app の責務なので
 * `@/app/store` にある。一方それを使う側（pages / widgets / features）は app を
 * import できない（下向きの一方通行）。そこで「RootState を知っている型付き hook」を
 * shared に置き、皆はこれを使う。shared が app を参照するのは FSD 公式が認める
 * 「app と shared は相互 import 可」の例外にあたる。
 */
export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;
