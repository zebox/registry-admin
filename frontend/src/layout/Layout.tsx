import * as React from 'react';
import { Layout, LayoutProps } from 'react-admin';
import AppBar from './AppBar';
import Menu from './Menu';

// eslint-disable-next-line 
export default (props: LayoutProps) => ( 
     <Layout {...props} appBar={AppBar} menu={Menu} />
);