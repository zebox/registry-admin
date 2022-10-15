
import { TextInput, useTranslate } from 'react-admin';

export const SearchFieldTranslated = (additionalComponent?: any[]): any => {
    const translate = useTranslate();
    if (additionalComponent && additionalComponent.length > 0) {
        return [
            <TextInput source="q" label={translate('ra.action.search')} alwaysOn />,
            { ...additionalComponent }
        ]
    }
    return [
        <TextInput source="q" label={translate('ra.action.search')} alwaysOn />
    ]

}
