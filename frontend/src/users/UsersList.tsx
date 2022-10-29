
import {
    List,
    Datagrid,
    TextField,
    EditButton,
    DeleteButton,
    SelectInput,
    TextInput,
    ReferenceField,
    useRecordContext,
    useCreatePath

} from 'react-admin';
import Button from '@mui/material/Button';
import { Link } from 'react-router-dom';
import HttpsIcon from '@mui/icons-material/Https';

import { DisabledField } from '../components/DisabledField';
import { SearchFieldTranslated } from '../helpers/Helpers'

interface IRoleList {
    id: string;
    name: string;
}

export const RoleList: Array<IRoleList> = [
    { id: 'user', name: 'User' },
    { id: 'manager', name: 'Manager' },
    { id: 'admin', name: 'Admin' }
]


const userFilters = [
    <TextInput source="q" label="Search" alwaysOn />,
    <SelectInput
        source="role"
        defaultValue={"user"}
        emptyValue={null}
        choices={RoleList} />
];

const AddRepositoryAccess =()=>{
    const record = useRecordContext();
    const createPath = useCreatePath();

    return record ? (
        <Button
            color="primary"
            component={Link}
            to={{
                pathname: '/access',
                search: `filter=${JSON.stringify({ owner_id: record.id })}`,
            }}
        >
            <HttpsIcon/>
        </Button>
    ) : null;

}

const UserList = () => (
    <List
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
        filters={SearchFieldTranslated([<SelectInput
            source="role"
            defaultValue={"user"}
            emptyValue={null}
            choices={RoleList} />])}
    >
        <Datagrid bulkActionButtons={false}>
            <TextField source="id" />
            <TextField source="login" />
            <TextField source="name" />
            <ReferenceField source="group" reference="groups">
                <TextField source="name" />
            </ReferenceField>
            <TextField source="role" />
            <DisabledField source="blocked" />
            <AddRepositoryAccess/>
            <EditButton />
            <DeleteButton />
        </Datagrid>
    </List>
);


export default UserList;