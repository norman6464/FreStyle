import { createSlice } from '@reduxjs/toolkit';

const initialState = {
  accessToken: localStorage.getItem('accessToken') || null,
  sub: localStorage.getItem('sub') || null,
  name: localStorage.getItem('name') || null,
  email: localStorage.getItem('email') || null,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setAuthData(state, action) {
      const { accessToken, sub, name, email } = action.payload;

      state.accessToken = accessToken;
      state.sub = sub;
      state.name = name;
      state.email = email;

      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('sub', sub);
      localStorage.setItem('name', name);
      localStorage.setItem('email', email);
    },

    clearAuthData(state) {
      state.accessToken = null;
      state.sub = null;
      state.name = null;
      state.email = null;

      localStorage.removeItem('accessToken');
      localStorage.removeItem('sub');
      localStorage.removeItem('name');
      localStorage.removeItem('email');
    },
  },
});

export const { setAuthData, clearAuthData } = authSlice.actions;
export default authSlice.reducer;
