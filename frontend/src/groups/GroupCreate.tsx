import {Create, TextInput, SimpleForm, useTranslate, usePermissions, NotFound} from 'react-admin';
import {requirePermission} from '../helpers/Helpers';

export const GroupCreate = () => {
    const translate = useTranslate();
    const {permissions} = usePermissions();

    return (requirePermission(permissions, 'admin') ?
            <Create title={translate('resources.groups.add_title')}>
                <SimpleForm>
                    <TextInput source="name" autoComplete="new-name"/>
                    <TextInput source="description" autoComplete='off' fullWidth/>
                </SimpleForm>
            </Create> : <NotFound/>
    )
};

export default GroupCreate;