
import * as React from "react";
import {
    List,
    Datagrid,
    TextField,
    EditButton,
    DeleteButton,
    usePermissions,
    NotFound
} from 'react-admin';
import {requirePermission} from '../helpers/Helpers';

const GroupList = () => {
    const { permissions } = usePermissions();
    return (
        requirePermission(permissions, 'admin') ?
            <List
                sort={{ field: 'name', order: 'ASC' }}
                perPage={25}
            >
                <Datagrid bulkActionButtons={false}
                    sx={{
                        '& .column-name': { width: '80%' },
                    }}>
                    <TextField source="id" />
                    <TextField source="name" />
                    <EditButton />
                    <DeleteButton />
                </Datagrid>
            </List> : <NotFound />
    )
};


export default GroupList;

function makeStyles(arg0: { toolbar: { alignItems: string; display: string; marginTop: number; marginBottom: number; }; }) {
    throw new Error("Function not implemented.");
}
