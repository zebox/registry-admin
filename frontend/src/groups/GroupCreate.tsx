
import { Create, TextInput, SimpleForm,useTranslate } from 'react-admin';
export const GroupCreate = () => {
    const translate = useTranslate();
    return (
        <Create title={translate('resources.groups.add_title')} >
            <SimpleForm>
                <TextInput source="name" autoComplete="new-name" />
                <TextInput source="description" autoComplete='off' fullWidth />
            </SimpleForm>

        </Create>
    )
};

export default GroupCreate;