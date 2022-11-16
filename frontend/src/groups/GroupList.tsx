
import * as React from "react";
import {
    List,
    Datagrid,
    TextField,
    EditButton,
    DeleteButton,
} from 'react-admin';

const GroupList = () => (
    <List
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
    >
        <Datagrid bulkActionButtons={false}>
                <TextField source="id" />
                <TextField source="name"  />
                <EditButton />
                <DeleteButton  />
        </Datagrid>
    </List>
);


export default GroupList;

function makeStyles(arg0: { toolbar: { alignItems: string; display: string; marginTop: number; marginBottom: number; }; }) {
    throw new Error("Function not implemented.");
}
