import { ChangeEvent } from 'react';

/** InputField の onChange（ChangeEvent 必須）を setter に橋渡しする。 */
export function toChangeHandler(setter: (value: string) => void) {
  return (e: ChangeEvent<HTMLInputElement>) => setter(e.target.value);
}
