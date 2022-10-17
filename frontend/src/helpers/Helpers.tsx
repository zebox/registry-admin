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

export const ConverUnixTimeToDate = (unixTimestamp:number):string =>{
    var date = new Date(unixTimestamp*1000);
    var hours = "0"+date.getHours();
    var minutes = "0"+date.getMinutes()
    var seconds = "0"+date.getSeconds()
    return "Date: "+date.getDate()+
              "/"+(date.getMonth()+1)+
              "/"+date.getFullYear()+
              " "+hours+
              ":"+minutes+
              ":"+seconds;

}