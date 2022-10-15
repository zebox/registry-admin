import { ReactElement } from 'react'; 
import { TextInput, useTranslate } from 'react-admin';

export const SearchFieldTranslated = (additionalComponent?:ReactElement<any, any>[]): any => {
    const translate = useTranslate();
    let filters = [
        <TextInput source="q" label={translate('ra.action.search')} alwaysOn />
    ]
    if (additionalComponent && additionalComponent.length > 0) {
        additionalComponent.map(element=>{
            filters.push(element);
        })
    }
    return filters;

}
