
import * as React from "react";
import {
    List,
    Datagrid,
    TextField,
    EditButton,
    useTranslate,
    usePermissions,
    NotFound
} from 'react-admin';
import { DeleteCustomButtonWithConfirmation } from "../components/DeleteCustomButtonWithConfirmation";
import { requirePermission, SearchFieldTranslated } from '../helpers/Helpers';

const GroupList = (props: any) => {
    const translate = useTranslate();
    const { permissions } = usePermissions();

    return (
        requirePermission(permissions, 'admin') ?
            <List
                sort={{ field: 'name', order: 'ASC' }}
                perPage={25}
                filters={SearchFieldTranslated(translate)}
            >
                <Datagrid bulkActionButtons={false}
                    sx={{
                        '& .column-name': { width: '80%' },
                    }}>
                    <TextField source="id" />
                    <TextField source="name" />
                    <EditButton />
                    <DeleteCustomButtonWithConfirmation source="name" {...props} />
                </Datagrid>
            </List> : <NotFound />
    )
};


export default GroupList;
